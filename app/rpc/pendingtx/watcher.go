package pendingtx

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	rpcfilters "github.com/FiboChain/fbc/app/rpc/namespaces/eth/filters"
	rpctypes "github.com/FiboChain/fbc/app/rpc/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/client/context"
	"github.com/FiboChain/fbc/libs/tendermint/libs/log"
	coretypes "github.com/FiboChain/fbc/libs/tendermint/rpc/core/types"
	tmtypes "github.com/FiboChain/fbc/libs/tendermint/types"
	"github.com/FiboChain/fbc/x/evm/watcher"
)

type Watcher struct {
	clientCtx context.CLIContext
	events    *rpcfilters.EventSystem
	logger    log.Logger

	sender Sender
}

type Sender interface {
	SendPending(hash []byte, tx *watcher.Transaction) error
	SendRmPending(hash []byte, tx *RmPendingTx) error
}

func NewWatcher(clientCtx context.CLIContext, log log.Logger, sender Sender) *Watcher {
	return &Watcher{
		clientCtx: clientCtx,
		events:    rpcfilters.NewEventSystem(clientCtx.Client),
		logger:    log.With("module", "pendingtx-watcher"),

		sender: sender,
	}
}

func (w *Watcher) Start() {
	pendingSub, _, err := w.events.SubscribePendingTxs()
	if err != nil {
		w.logger.Error("error creating block filter", "error", err.Error())
	}

	rmPendingSub, _, err := w.events.SubscribeRmPendingTx()
	if err != nil {
		w.logger.Error("error creating block filter", "error", err.Error())
	}

	go func(pendingCh <-chan coretypes.ResultEvent, rmPendingdCh <-chan coretypes.ResultEvent) {
		for {
			select {
			case re := <-pendingCh:
				data, ok := re.Data.(tmtypes.EventDataTx)
				if !ok {
					w.logger.Error(fmt.Sprintf("invalid pending tx data type %T, expected EventDataTx", re.Data))
					continue
				}
				txHash := common.BytesToHash(data.Tx.Hash(data.Height))
				w.logger.Debug("receive pending tx", "txHash=", txHash.String())

				// only watch evm tx
				ethTx, err := rpctypes.RawTxToEthTx(w.clientCtx, data.Tx, data.Height)
				if err != nil {
					w.logger.Error("failed to decode raw tx to eth tx", "hash", txHash.String(), "error", err)
					continue
				}

				tx, err := watcher.NewTransaction(ethTx, txHash, common.Hash{}, uint64(data.Height), uint64(data.Index))
				if err != nil {
					w.logger.Error("failed to new transaction", "hash", txHash.String(), "error", err)
					continue
				}

				go func() {
					w.logger.Debug("push pending tx to MQ", "txHash=", tx.Hash.String())
					err = w.sender.SendPending(tx.Hash.Bytes(), tx)
					if err != nil {
						w.logger.Error("failed to send pending tx", "hash", tx.Hash.String(), "error", err)
					}
				}()
			case re := <-rmPendingdCh:
				data, ok := re.Data.(tmtypes.EventDataRmPendingTx)
				if !ok {
					w.logger.Error(fmt.Sprintf("invalid rm pending tx data type %T, expected EventDataTx", re.Data))
					continue
				}
				txHash := common.BytesToHash(data.Hash).String()
				go func() {
					w.logger.Debug("push rm pending tx to MQ", "txHash=", txHash)
					err = w.sender.SendRmPending(data.Hash, &RmPendingTx{
						From:   data.From,
						Hash:   txHash,
						Nonce:  hexutil.Uint64(data.Nonce).String(),
						Delete: true,
						Reason: data.Reason,
					})
					if err != nil {
						w.logger.Error("failed to send rm pending tx", "hash", txHash, "error", err)
					}
				}()
			}
		}
	}(pendingSub.Event(), rmPendingSub.Event())
}
