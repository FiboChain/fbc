package v0

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	amino "github.com/tendermint/go-amino"

	"github.com/FiboChain/fbc/libs/tendermint/libs/log"
	"github.com/FiboChain/fbc/libs/tendermint/p2p"
	sm "github.com/FiboChain/fbc/libs/tendermint/state"
	"github.com/FiboChain/fbc/libs/tendermint/store"
	"github.com/FiboChain/fbc/libs/tendermint/types"
)

const (
	// BlockchainChannel is a channel for blocks and status updates (`BlockStore` height)
	BlockchainChannel = byte(0x40)

	trySyncIntervalMS = 10

	// stop syncing when last block's time is
	// within this much of the system time.
	// stopSyncingDurationMinutes = 10

	// ask for best height every 10s
	statusUpdateIntervalSeconds = 10
	// check if we should switch to consensus reactor
	switchToConsensusIntervalSeconds = 1

	// NOTE: keep up to date with bcBlockResponseMessage
	bcBlockResponseMessagePrefixSize   = 4
	bcBlockResponseMessageFieldKeySize = 1
	maxMsgSize                         = types.MaxBlockSizeBytes +
		bcBlockResponseMessagePrefixSize +
		bcBlockResponseMessageFieldKeySize

	maxIntervalForFastSync        = 3
	maxPeersProportionForFastSync = 0.4
)

type consensusReactor interface {
	// SwitchToConsensus called when we switch from blockchain reactor and fast sync to
	// the consensus machine
	SwitchToConsensus(sm.State, uint64) bool

	// SwitchToFastSync called when we switch from the consensus machine to blockchain reactor and fast sync
	SwitchToFastSync() (sm.State, error)
}

type peerError struct {
	err    error
	peerID p2p.ID
}

func (e peerError) Error() string {
	return fmt.Sprintf("error with peer %v: %s", e.peerID, e.err.Error())
}

// BlockchainReactor handles long-term catchup syncing.
type BlockchainReactor struct {
	p2p.BaseReactor

	// mutable
	curState sm.State

	blockExec    *sm.BlockExecutor
	store        *store.BlockStore
	dstore       *store.DeltaStore
	pool         *BlockPool
	fastSync     bool
	autoFastSync bool
	isSyncing    bool
	mtx          sync.RWMutex

	requestsCh <-chan BlockRequest
	errorsCh   <-chan peerError
}

// NewBlockchainReactor returns new reactor instance.
func NewBlockchainReactor(state sm.State, blockExec *sm.BlockExecutor, store *store.BlockStore, dstore *store.DeltaStore,
	fastSync bool) *BlockchainReactor {
	if state.LastBlockHeight != store.Height() {
		panic(fmt.Sprintf("state (%v) and store (%v) height mismatch", state.LastBlockHeight,
			store.Height()))
	}

	requestsCh := make(chan BlockRequest, maxTotalRequesters)

	const capacity = 1000                      // must be bigger than peers count
	errorsCh := make(chan peerError, capacity) // so we don't block in #Receive#pool.AddBlock

	pool := NewBlockPool(
		store.Height()+1,
		requestsCh,
		errorsCh,
	)

	bcR := &BlockchainReactor{
		curState:   state,
		blockExec:  blockExec,
		store:      store,
		dstore:     dstore,
		pool:       pool,
		fastSync:   fastSync,
		mtx:        sync.RWMutex{},
		requestsCh: requestsCh,
		errorsCh:   errorsCh,
	}
	bcR.BaseReactor = *p2p.NewBaseReactor("BlockchainReactor", bcR)
	return bcR
}

// SetLogger implements service.Service by setting the logger on reactor and pool.
func (bcR *BlockchainReactor) SetLogger(l log.Logger) {
	bcR.BaseService.Logger = l
	bcR.pool.Logger = l
}

// OnStart implements service.Service.
func (bcR *BlockchainReactor) OnStart() error {
	if bcR.fastSync {
		err := bcR.pool.Start()
		if err != nil {
			return err
		}
		go bcR.poolRoutine()
	}
	return nil
}

// OnStop implements service.Service.
func (bcR *BlockchainReactor) OnStop() {
	bcR.pool.Stop()
}

// GetChannels implements Reactor
func (bcR *BlockchainReactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		{
			ID:                  BlockchainChannel,
			Priority:            10,
			SendQueueCapacity:   1000,
			RecvBufferCapacity:  50 * 4096,
			RecvMessageCapacity: maxMsgSize,
		},
	}
}

// AddPeer implements Reactor by sending our state to peer.
func (bcR *BlockchainReactor) AddPeer(peer p2p.Peer) {
	msgBytes := cdc.MustMarshalBinaryBare(&bcStatusResponseMessage{
		Height: bcR.store.Height(),
		Base:   bcR.store.Base(),
	})
	peer.Send(BlockchainChannel, msgBytes)
	// it's OK if send fails. will try later in poolRoutine

	// peer is added to the pool once we receive the first
	// bcStatusResponseMessage from the peer and call pool.SetPeerRange
}

// RemovePeer implements Reactor by removing peer from the pool.
func (bcR *BlockchainReactor) RemovePeer(peer p2p.Peer, reason interface{}) {
	bcR.pool.RemovePeer(peer.ID())
}

// respondToPeer loads a block and sends it to the requesting peer,
// if we have it. Otherwise, we'll respond saying we don't have it.
func (bcR *BlockchainReactor) respondToPeer(msg *bcBlockRequestMessage,
	src p2p.Peer) (queued bool) {

	block := bcR.store.LoadBlock(msg.Height)
	if block != nil {
		msgBytes := cdc.MustMarshalBinaryBare(&bcBlockResponseMessage{Block: block})
		return src.TrySend(BlockchainChannel, msgBytes)
	}

	bcR.Logger.Info("Peer asking for a block we don't have", "src", src, "height", msg.Height)

	msgBytes := cdc.MustMarshalBinaryBare(&bcNoBlockResponseMessage{Height: msg.Height})
	return src.TrySend(BlockchainChannel, msgBytes)
}

// Receive implements Reactor by handling 4 types of messages (look below).
func (bcR *BlockchainReactor) Receive(chID byte, src p2p.Peer, msgBytes []byte) {
	msg, err := decodeMsg(msgBytes)
	if err != nil {
		bcR.Logger.Error("Error decoding message", "src", src, "chId", chID, "msg", msg, "err", err, "bytes", msgBytes)
		bcR.Switch.StopPeerForError(src, err)
		return
	}

	if err = msg.ValidateBasic(); err != nil {
		bcR.Logger.Error("Peer sent us invalid msg", "peer", src, "msg", msg, "err", err)
		bcR.Switch.StopPeerForError(src, err)
		return
	}

	bcR.Logger.Debug("Receive", "src", src, "chID", chID, "msg", msg)

	switch msg := msg.(type) {
	case *bcBlockRequestMessage:
		bcR.respondToPeer(msg, src)
	case *bcBlockResponseMessage:
		bcR.Logger.Info("AddBlock.", "Height", msg.Block.Height, "Peer", src.ID())
		bcR.pool.AddBlock(src.ID(), msg.Block, msg.Deltas, len(msgBytes))
	case *bcStatusRequestMessage:
		// Send peer our state.
		src.TrySend(BlockchainChannel, cdc.MustMarshalBinaryBare(&bcStatusResponseMessage{
			Height: bcR.store.Height(),
			Base:   bcR.store.Base(),
		}))
	case *bcStatusResponseMessage:
		// Got a peer status. Unverified. TODO: should verify before SetPeerRange
		shouldSync := bcR.pool.SetPeerRange(src.ID(), msg.Base, msg.Height, bcR.store.Height())
		// should switch to fast-sync when more than XX peers' height is greater than store.Height
		if shouldSync {
			bcR.Logger.Info("ShouldSync.", "Status peer", msg.Height, "now", bcR.store.Height())
			go bcR.poolRoutine()
		}
	case *bcNoBlockResponseMessage:
		bcR.Logger.Debug("Peer does not have requested block", "peer", src, "height", msg.Height)
	default:
		bcR.Logger.Error(fmt.Sprintf("Unknown message type %v", reflect.TypeOf(msg)))
	}
}

// Handle messages from the poolReactor telling the reactor what to do.
// NOTE: Don't sleep in the FOR_LOOP or otherwise slow it down!
func (bcR *BlockchainReactor) poolRoutine() {
	bcR.mtx.Lock()
	if bcR.isSyncing {
		bcR.mtx.Unlock()
		return
	}
	bcR.isSyncing = true
	bcR.mtx.Unlock()

	defer func() {
		bcR.setIsSyncing(false)
	}()

	conR, ok := bcR.Switch.Reactor("CONSENSUS").(consensusReactor)
	if ok {
		conState, err := conR.SwitchToFastSync()
		if err == nil {
			bcR.curState = conState
		}
	}
	chainID := bcR.curState.ChainID

	bcR.pool.SetHeight(bcR.curState.LastBlockHeight + 1)
	bcR.pool.Stop()
	bcR.pool.Reset()
	bcR.pool.Start()

	blocksSynced := uint64(0)

	lastHundred := time.Now()
	lastRate := 0.0

	switchToConsensusTicker := time.NewTicker(switchToConsensusIntervalSeconds * time.Second)
	trySyncTicker := time.NewTicker(trySyncIntervalMS * time.Millisecond)
	statusUpdateTicker := time.NewTicker(statusUpdateIntervalSeconds * time.Second)

	didProcessCh := make(chan struct{}, 1)

	go func() {
		for {
			select {
			case <-bcR.Quit():
				return
			case <-bcR.pool.Quit():
				return
			case request := <-bcR.requestsCh:
				peer := bcR.Switch.Peers().Get(request.PeerID)
				if peer == nil {
					continue
				}
				msgBytes := cdc.MustMarshalBinaryBare(&bcBlockRequestMessage{request.Height})
				queued := peer.TrySend(BlockchainChannel, msgBytes)
				if !queued {
					bcR.Logger.Debug("Send queue is full, drop block request", "peer", peer.ID(), "height", request.Height)
				}
			case err := <-bcR.errorsCh:
				peer := bcR.Switch.Peers().Get(err.peerID)
				if peer != nil {
					bcR.Switch.StopPeerForError(peer, err)
				}

			case <-statusUpdateTicker.C:
				// ask for status updates
				go bcR.BroadcastStatusRequest() // nolint: errcheck
			}
		}
	}()

FOR_LOOP:
	for {
		select {
		case <-switchToConsensusTicker.C:
			if bcR.SwitchToConsensus(bcR.curState) {
				break FOR_LOOP
			}

		case <-trySyncTicker.C: // chan time
			select {
			case didProcessCh <- struct{}{}:
			default:
			}

		case <-didProcessCh:
			// NOTE: It is a subtle mistake to process more than a single block
			// at a time (e.g. 10) here, because we only TrySend 1 request per
			// loop.  The ratio mismatch can result in starving of blocks, a
			// sudden burst of requests and responses, and repeat.
			// Consequently, it is better to split these routines rather than
			// coupling them as it's written here.  TODO uncouple from request
			// routine.

			// See if there are any blocks to sync.
			first, second, _ := bcR.pool.PeekTwoBlocks()
			//bcR.Logger.Info("TrySync peeked", "first", first, "second", second)
			if first == nil || second == nil {
				// We need both to sync the first block.
				continue FOR_LOOP
			} else {
				// Try again quickly next loop.
				didProcessCh <- struct{}{}
			}
			bcR.Logger.Info("PeekTwoBlocks.", "First", first.Height, "Second", second.Height)

			firstParts := first.MakePartSet(types.BlockPartSizeBytes)
			firstPartsHeader := firstParts.Header()
			firstID := types.BlockID{Hash: first.Hash(), PartsHeader: firstPartsHeader}
			// Finally, verify the first block using the second's commit
			// NOTE: we can probably make this more efficient, but note that calling
			// first.Hash() doesn't verify the tx contents, so MakePartSet() is
			// currently necessary.
			err := bcR.curState.Validators.VerifyCommit(
				chainID, firstID, first.Height, second.LastCommit)
			if err != nil {
				bcR.Logger.Error("Error in validation", "err", err)
				peerID := bcR.pool.RedoRequest(first.Height)
				peer := bcR.Switch.Peers().Get(peerID)
				if peer != nil {
					// NOTE: we've already removed the peer's request, but we
					// still need to clean up the rest.
					bcR.Switch.StopPeerForError(peer, fmt.Errorf("blockchainReactor validation error: %v", err))
				}
				peerID2 := bcR.pool.RedoRequest(second.Height)
				peer2 := bcR.Switch.Peers().Get(peerID2)
				if peer2 != nil && peer2 != peer {
					// NOTE: we've already removed the peer's request, but we
					// still need to clean up the rest.
					bcR.Switch.StopPeerForError(peer2, fmt.Errorf("blockchainReactor validation error: %v", err))
				}
				continue FOR_LOOP
			} else {
				bcR.pool.PopRequest()

				// TODO: batch saves so we dont persist to disk every block
				bcR.store.SaveBlock(first, firstParts, second.LastCommit)

				// TODO: same thing for app - but we would need a way to
				// get the hash without persisting the state
				var err error
				bcR.curState, _, err = bcR.blockExec.ApplyBlock(bcR.curState, firstID, first) // rpc
				if err != nil {
					// TODO This is bad, are we zombie?
					// The block can't be committed, do we need to delete it from store db?
					_, errDel := bcR.store.DeleteBlocksFromTop(first.Height - 1)
					if errDel != nil {
						bcR.Logger.Error("Failed to delete blocks from top", "height", first.Height-1, "err", errDel)
					}
					panic(fmt.Sprintf("Failed to process committed block (%d:%X): %v", first.Height, first.Hash(), err))
				}
				blocksSynced++

				/*
					if types.EnableBroadcastP2PDelta() {
						// persists the given deltas to the underlying db.
						deltas.Height = first.Height
						bcR.dstore.SaveDeltas(deltas, first.Height)
					}
				*/

				if blocksSynced%100 == 0 {
					lastRate = 0.9*lastRate + 0.1*(100/time.Since(lastHundred).Seconds())
					bcR.Logger.Info("Fast Sync Rate", "height", bcR.pool.height,
						"max_peer_height", bcR.pool.MaxPeerHeight(), "blocks/s", lastRate)
					lastHundred = time.Now()
				}
			}
			continue FOR_LOOP

		case <-bcR.Quit():
			break FOR_LOOP
		}
	}
}

func (bcR *BlockchainReactor) CheckFastSyncCondition() {
	// ask for status updates
	bcR.Logger.Info("CheckFastSyncCondition.")
	go bcR.BroadcastStatusRequest()
}

func (bcR *BlockchainReactor) SwitchToConsensus(state sm.State) bool {
	if !bcR.getIsSyncing() {
		return false
	}

	blocksSynced := uint64(0)
	height, numPending, lenRequesters := bcR.pool.GetStatus()
	outbound, inbound, _ := bcR.Switch.NumPeers()
	bcR.Logger.Debug("Consensus ticker", "numPending", numPending, "total", lenRequesters,
		"outbound", outbound, "inbound", inbound)
	conR, ok := bcR.Switch.Reactor("CONSENSUS").(consensusReactor)
	if bcR.pool.IsCaughtUp() && ok {
		bcR.Logger.Info("Time to switch to consensus reactor!", "height", height)

		succeed := conR.SwitchToConsensus(state, blocksSynced)
		if succeed {
			bcR.pool.Stop()
			return true
		}
	}
	return false
}

// BroadcastStatusRequest broadcasts `BlockStore` base and height.
func (bcR *BlockchainReactor) BroadcastStatusRequest() error {
	msgBytes := cdc.MustMarshalBinaryBare(&bcStatusRequestMessage{
		Base:   bcR.store.Base(),
		Height: bcR.store.Height(),
	})
	bcR.Switch.Broadcast(BlockchainChannel, msgBytes)
	return nil
}

func (bcR *BlockchainReactor) setIsSyncing(value bool) {
	bcR.mtx.Lock()
	bcR.isSyncing = value
	bcR.mtx.Unlock()
}

func (bcR *BlockchainReactor) getIsSyncing() bool {
	bcR.mtx.Lock()
	defer bcR.mtx.Unlock()
	return bcR.isSyncing
}

//-----------------------------------------------------------------------------
// Messages

// BlockchainMessage is a generic message for this reactor.
type BlockchainMessage interface {
	ValidateBasic() error
}

// RegisterBlockchainMessages registers the fast sync messages for amino encoding.
func RegisterBlockchainMessages(cdc *amino.Codec) {
	cdc.RegisterInterface((*BlockchainMessage)(nil), nil)
	cdc.RegisterConcrete(&bcBlockRequestMessage{}, "tendermint/blockchain/BlockRequest", nil)
	cdc.RegisterConcrete(&bcBlockResponseMessage{}, "tendermint/blockchain/BlockResponse", nil)
	cdc.RegisterConcrete(&bcNoBlockResponseMessage{}, "tendermint/blockchain/NoBlockResponse", nil)
	cdc.RegisterConcrete(&bcStatusResponseMessage{}, "tendermint/blockchain/StatusResponse", nil)
	cdc.RegisterConcrete(&bcStatusRequestMessage{}, "tendermint/blockchain/StatusRequest", nil)
}

func decodeMsg(bz []byte) (msg BlockchainMessage, err error) {
	if len(bz) > maxMsgSize {
		return msg, fmt.Errorf("msg exceeds max size (%d > %d)", len(bz), maxMsgSize)
	}
	err = cdc.UnmarshalBinaryBare(bz, &msg)
	return
}

//-------------------------------------

type bcBlockRequestMessage struct {
	Height int64
}

// ValidateBasic performs basic validation.
func (m *bcBlockRequestMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative Height")
	}
	return nil
}

func (m *bcBlockRequestMessage) String() string {
	return fmt.Sprintf("[bcBlockRequestMessage %v]", m.Height)
}

type bcNoBlockResponseMessage struct {
	Height int64
}

// ValidateBasic performs basic validation.
func (m *bcNoBlockResponseMessage) ValidateBasic() error {
	if m.Height < 0 {
		return errors.New("negative Height")
	}
	return nil
}

func (m *bcNoBlockResponseMessage) String() string {
	return fmt.Sprintf("[bcNoBlockResponseMessage %d]", m.Height)
}

//-------------------------------------

type bcBlockResponseMessage struct {
	Block  *types.Block
	Deltas *types.Deltas
}

// ValidateBasic performs basic validation.
func (m *bcBlockResponseMessage) ValidateBasic() error {
	return m.Block.ValidateBasic()
}

func (m *bcBlockResponseMessage) String() string {
	return fmt.Sprintf("[bcBlockResponseMessage %v]", m.Block.Height)
}

//-------------------------------------

type bcStatusRequestMessage struct {
	Height int64
	Base   int64
}

// ValidateBasic performs basic validation.
func (m *bcStatusRequestMessage) ValidateBasic() error {
	if m.Base < 0 {
		return errors.New("negative Base")
	}
	if m.Height < 0 {
		return errors.New("negative Height")
	}
	if m.Base > m.Height {
		return fmt.Errorf("base %v cannot be greater than height %v", m.Base, m.Height)
	}
	return nil
}

func (m *bcStatusRequestMessage) String() string {
	return fmt.Sprintf("[bcStatusRequestMessage %v:%v]", m.Base, m.Height)
}

//-------------------------------------

type bcStatusResponseMessage struct {
	Height int64
	Base   int64
}

// ValidateBasic performs basic validation.
func (m *bcStatusResponseMessage) ValidateBasic() error {
	if m.Base < 0 {
		return errors.New("negative Base")
	}
	if m.Height < 0 {
		return errors.New("negative Height")
	}
	if m.Base > m.Height {
		return fmt.Errorf("base %v cannot be greater than height %v", m.Base, m.Height)
	}
	return nil
}

func (m *bcStatusResponseMessage) String() string {
	return fmt.Sprintf("[bcStatusResponseMessage %v:%v]", m.Base, m.Height)
}