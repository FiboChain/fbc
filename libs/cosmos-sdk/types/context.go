package types

import (
	"context"
	"time"

	"github.com/FiboChain/fbc/libs/tendermint/trace"

	abci "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	"github.com/FiboChain/fbc/libs/tendermint/libs/log"
	"github.com/gogo/protobuf/proto"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/store/gaskv"
	stypes "github.com/FiboChain/fbc/libs/cosmos-sdk/store/types"
)

/*
Context is an immutable object contains all information needed to
process a request.

It contains a context.Context object inside if you want to use that,
but please do not over-use it. We try to keep all data structured
and standard additions here would be better just to add to the Context struct
*/
type Context struct {
	ctx            context.Context
	ms             MultiStore
	header         *abci.Header
	chainID        string
	from           string
	txBytes        []byte
	logger         log.Logger
	voteInfo       []abci.VoteInfo
	gasMeter       GasMeter
	blockGasMeter  GasMeter
	isDeliver      bool
	checkTx        bool
	recheckTx      bool // if recheckTx == true, then checkTx must also be true
	wrappedCheckTx bool // if wrappedCheckTx == true, then checkTx must also be true
	traceTx        bool // traceTx is set true for trace tx and its predesessors , traceTx was set in app.beginBlockForTrace()
	traceTxLog     bool // traceTxLog is used to create trace logger for evm , traceTxLog is set to true when only tracing target tx (its predesessors will set false), traceTxLog is set before runtx
	isAsync        bool
	minGasPrice    DecCoins
	consParams     *abci.ConsensusParams
	eventManager   *EventManager
	accountNonce   uint64
	cache          *Cache
	trc            *trace.Tracer
	accountCache   *AccountCache
}

// Proposed rename, not done to avoid API breakage
type Request = Context

// Read-only accessors
func (c *Context) Context() context.Context { return c.ctx }
func (c *Context) MultiStore() MultiStore   { return c.ms }
func (c *Context) BlockHeight() int64 {
	if c.header == nil {
		return 0
	}
	return c.header.Height
}
func (c *Context) BlockTime() time.Time {
	if c.header == nil {
		return time.Time{}
	}
	return c.header.Time
}
func (c *Context) ChainID() string            { return c.chainID }
func (c *Context) From() string               { return c.from }
func (c *Context) TxBytes() []byte            { return c.txBytes }
func (c *Context) Logger() log.Logger         { return c.logger }
func (c *Context) VoteInfos() []abci.VoteInfo { return c.voteInfo }
func (c *Context) GasMeter() GasMeter         { return c.gasMeter }
func (c *Context) BlockGasMeter() GasMeter    { return c.blockGasMeter }
func (c *Context) IsDeliver() bool {
	return c.isDeliver
}
func (c *Context) IsCheckTx() bool             { return c.checkTx }
func (c *Context) IsReCheckTx() bool           { return c.recheckTx }
func (c *Context) IsTraceTx() bool             { return c.traceTx }
func (c *Context) IsTraceTxLog() bool          { return c.traceTxLog }
func (c *Context) IsWrappedCheckTx() bool      { return c.wrappedCheckTx }
func (c *Context) MinGasPrices() DecCoins      { return c.minGasPrice }
func (c *Context) EventManager() *EventManager { return c.eventManager }
func (c *Context) IsAsync() bool               { return c.isAsync }
func (c *Context) AccountNonce() uint64        { return c.accountNonce }
func (c *Context) AnteTracer() *trace.Tracer   { return c.trc }
func (c *Context) Cache() *Cache {
	return c.cache
}

type AccountCache struct {
	FromAcc       interface{} // must be auth.Account
	ToAcc         interface{} // must be auth.Account
	FromAccGotGas Gas
	ToAccGotGas   Gas
}

func (c *Context) EnableAccountCache()  { c.accountCache = &AccountCache{} }
func (c *Context) DisableAccountCache() { c.accountCache = nil }

func (c *Context) GetFromAccountCacheData() interface{} {
	if c.accountCache == nil {
		return nil
	}
	return c.accountCache.FromAcc
}

func (c *Context) GetFromAccountCacheGas() Gas {
	if c.accountCache == nil {
		return 0
	}
	return c.accountCache.FromAccGotGas
}

func (c *Context) GetToAccountCacheData() interface{} {
	if c.accountCache == nil {
		return nil
	}
	return c.accountCache.ToAcc
}

func (c *Context) GetToAccountCacheGas() Gas {
	if c.accountCache == nil {
		return 0
	}
	return c.accountCache.ToAccGotGas
}

func (c *Context) UpdateFromAccountCache(fromAcc interface{}, fromAccGettedGas Gas) {
	if c.accountCache != nil {
		c.accountCache.FromAcc = fromAcc
		c.accountCache.FromAccGotGas = fromAccGettedGas
	}
}

func (c *Context) UpdateToAccountCache(toAcc interface{}, toAccGotGas Gas) {
	if c.accountCache != nil {
		c.accountCache.ToAcc = toAcc
		c.accountCache.ToAccGotGas = toAccGotGas
	}
}

func (c *Context) BlockProposerAddress() []byte {
	if c.header == nil {
		return nil
	}
	return c.header.ProposerAddress
}

// clone the header before returning
func (c Context) BlockHeader() abci.Header {
	if c.header == nil {
		return abci.Header{}
	}
	var msg = proto.Clone(c.header).(*abci.Header)
	return *msg
}

func (c Context) ConsensusParams() *abci.ConsensusParams {
	return proto.Clone(c.consParams).(*abci.ConsensusParams)
}

// create a new context
func NewContext(ms MultiStore, header abci.Header, isCheckTx bool, logger log.Logger) Context {
	// https://github.com/gogo/protobuf/issues/519
	header.Time = header.Time.UTC()
	return Context{
		ctx:          context.Background(),
		ms:           ms,
		header:       &header,
		chainID:      header.ChainID,
		checkTx:      isCheckTx,
		logger:       logger,
		gasMeter:     stypes.NewInfiniteGasMeter(),
		minGasPrice:  DecCoins{},
		eventManager: NewEventManager(),
	}
}

func (c Context) WithContext(ctx context.Context) Context {
	c.ctx = ctx
	return c
}

func (c Context) WithMultiStore(ms MultiStore) Context {
	c.ms = ms
	return c
}

func (c Context) WithAsync() Context {
	c.isAsync = true
	return c
}

func (c *Context) SetDeliver() {
	c.isDeliver = true
}

func (c Context) WithBlockHeader(header abci.Header) Context {
	// https://github.com/gogo/protobuf/issues/519
	header.Time = header.Time.UTC()
	c.header = &header
	return c
}

func (c Context) WithBlockTime(newTime time.Time) Context {
	newHeader := c.BlockHeader()
	// https://github.com/gogo/protobuf/issues/519
	newHeader.Time = newTime.UTC()
	return c.WithBlockHeader(newHeader)
}

func (c Context) WithProposer(addr ConsAddress) Context {
	newHeader := c.BlockHeader()
	newHeader.ProposerAddress = addr.Bytes()
	return c.WithBlockHeader(newHeader)
}

func (c Context) WithBlockHeight(height int64) Context {
	newHeader := c.BlockHeader()
	newHeader.Height = height
	return c.WithBlockHeader(newHeader)
}

func (c Context) WithChainID(chainID string) Context {
	c.chainID = chainID
	return c
}

func (c Context) WithFrom(from string) Context {
	c.from = from
	return c
}

func (c Context) WithTxBytes(txBytes []byte) Context {
	c.txBytes = txBytes
	return c
}

func (c Context) WithLogger(logger log.Logger) Context {
	c.logger = logger
	return c
}

func (c Context) WithVoteInfos(voteInfo []abci.VoteInfo) Context {
	c.voteInfo = voteInfo
	return c
}

func (c Context) WithGasMeter(meter GasMeter) Context {
	c.gasMeter = meter
	return c
}

func (c Context) WithBlockGasMeter(meter GasMeter) Context {
	c.blockGasMeter = meter
	return c
}

func (c Context) WithIsCheckTx(isCheckTx bool) Context {
	c.checkTx = isCheckTx
	return c
}

// WithIsRecheckTx called with true will also set true on checkTx in order to
// enforce the invariant that if recheckTx = true then checkTx = true as well.
func (c Context) WithIsReCheckTx(isRecheckTx bool) Context {
	if isRecheckTx {
		c.checkTx = true
	}
	c.recheckTx = isRecheckTx
	return c
}
func (c Context) WithIsTraceTxLog(isTraceTxLog bool) Context {
	if isTraceTxLog {
		c.checkTx = true
	}
	c.traceTxLog = isTraceTxLog
	return c
}
func (c Context) WithIsTraceTx(isTraceTx bool) Context {
	if isTraceTx {
		c.checkTx = true
	}
	c.traceTx = isTraceTx
	return c
}

// WithIsWrappedCheckTx called with true will also set true on checkTx in order to
// enforce the invariant that if recheckTx = true then checkTx = true as well.
func (c Context) WithIsWrappedCheckTx(isWrappedCheckTx bool) Context {
	if isWrappedCheckTx {
		c.checkTx = true
	}
	c.wrappedCheckTx = isWrappedCheckTx
	return c
}

func (c Context) WithMinGasPrices(gasPrices DecCoins) Context {
	c.minGasPrice = gasPrices
	return c
}

func (c Context) WithConsensusParams(params *abci.ConsensusParams) Context {
	c.consParams = params
	return c
}

func (c Context) WithEventManager(em *EventManager) Context {
	c.eventManager = em
	return c
}

func (c Context) WithAccountNonce(nonce uint64) Context {
	c.accountNonce = nonce
	return c
}

func (c Context) WithCache(cache *Cache) Context {
	c.cache = cache
	return c
}

// TODO: remove???
func (c *Context) IsZero() bool {
	return c.ms == nil
}

// WithValue is deprecated, provided for backwards compatibility
// Please use
//     ctx = ctx.WithContext(context.WithValue(ctx.Context(), key, false))
// instead of
//     ctx = ctx.WithValue(key, false)
func (c Context) WithValue(key, value interface{}) Context {
	c.ctx = context.WithValue(c.ctx, key, value)
	return c
}

func (c *Context) SetGasMeter(meter GasMeter) {
	c.gasMeter = meter
}

func (c *Context) SetMultiStore(ms MultiStore) {
	c.ms = ms
}

func (c *Context) SetEventManager(em *EventManager) {
	c.eventManager = em
}

func (c *Context) SetAccountNonce(nonce uint64) {
	c.accountNonce = nonce
}

func (c *Context) SetCache(cache *Cache) {
	c.cache = cache
}

func (c *Context) SetFrom(from string) {
	c.from = from
}

// Value is deprecated, provided for backwards compatibility
// Please use
//     ctx.Context().Value(key)
// instead of
//     ctx.Value(key)
func (c Context) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

// ----------------------------------------------------------------------------
// Store / Caching
// ----------------------------------------------------------------------------

// KVStore fetches a KVStore from the MultiStore.
func (c *Context) KVStore(key StoreKey) KVStore {
	return gaskv.NewStore(c.MultiStore().GetKVStore(key), c.GasMeter(), stypes.KVGasConfig())
}

// TransientStore fetches a TransientStore from the MultiStore.
func (c Context) TransientStore(key StoreKey) KVStore {
	return gaskv.NewStore(c.MultiStore().GetKVStore(key), c.GasMeter(), stypes.TransientGasConfig())
}

// CacheContext returns a new Context with the multi-store cached and a new
// EventManager. The cached context is written to the context when writeCache
// is called.
func (c Context) CacheContext() (cc Context, writeCache func()) {
	cms := c.MultiStore().CacheMultiStore()
	cc = c.WithMultiStore(cms).WithEventManager(NewEventManager())
	return cc, cms.Write
}

// An emptyCtx  has no values. It is a
// struct{}.
func EmptyContext() Context {
	return Context{}
}

func (c Context) WithAnteTracer(trc *trace.Tracer) Context {
	c.trc = trc
	return c
}
