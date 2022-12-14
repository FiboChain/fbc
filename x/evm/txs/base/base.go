package base

import (
	"math/big"

	bam "github.com/FiboChain/fbc/libs/cosmos-sdk/baseapp"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	authexported "github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/exported"
	"github.com/FiboChain/fbc/x/common/analyzer"
	"github.com/FiboChain/fbc/x/evm/keeper"
	"github.com/FiboChain/fbc/x/evm/types"
)

// Keeper alias of keeper.Keeper, to solve import circle. also evm.Keeper is alias keeper.Keeper
type Keeper = keeper.Keeper

// Config tx's needed ctx and keeper
type Config struct {
	Ctx    sdk.Context
	Keeper *Keeper
}

// Result evm execute result
type Result struct {
	ExecResult     *types.ExecutionResult
	ResultData     *types.ResultData
	InnerTxs       interface{}
	Erc20Contracts interface{}
}

// Tx evm tx
type Tx struct {
	Ctx    sdk.Context
	Keeper *Keeper

	StateTransition types.StateTransition
}

// Prepare convert msg to state transition
func (tx *Tx) Prepare(msg *types.MsgEthereumTx) (err error) {
	tx.AnalyzeStart(bam.Txhash)
	defer tx.AnalyzeStop(bam.Txhash)

	tx.StateTransition, err = msg2st(&tx.Ctx, tx.Keeper, msg)
	return
}

// GetChainConfig get chain config, the chain config may cached
func (tx *Tx) GetChainConfig() (types.ChainConfig, bool) {
	return tx.Keeper.GetChainConfig(tx.Ctx)
}

// Transition execute evm tx
func (tx *Tx) Transition(config types.ChainConfig) (result Result, err error) {
	result.ExecResult, result.ResultData, err, result.InnerTxs, result.Erc20Contracts = tx.StateTransition.TransitionDb(tx.Ctx, config)

	// async mod goes immediately
	if tx.Ctx.IsAsync() {
		tx.Keeper.LogsManages.Set(string(tx.Ctx.TxBytes()), keeper.TxResult{
			ResultData: result.ResultData,
			Err:        err,
		})
	}

	return
}

// DecorateResult TraceTxLog situation Decorate the result
// it was replaced to trace logs when trace tx even if err != nil
func (tx *Tx) DecorateResult(inResult *Result, inErr error) (result *sdk.Result, err error) {
	if inErr != nil {
		return nil, inErr
	}
	return inResult.ExecResult.Result, inErr
}

func (tx *Tx) EmitEvent(msg *types.MsgEthereumTx, result *Result) {
	tx.Ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEthereumTx,
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Data.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, types.EthAddressStringer(tx.StateTransition.Sender).String()),
		),
	})

	if msg.Data.Recipient != nil {
		tx.Ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEthereumTx,
				sdk.NewAttribute(types.AttributeKeyRecipient, types.EthAddressStringer(*msg.Data.Recipient).String()),
			),
		)
	}

	// set the events to the result
	result.ExecResult.Result.Events = tx.Ctx.EventManager().Events()
}

func NewTx(config Config) *Tx {
	return &Tx{
		Ctx:    config.Ctx,
		Keeper: config.Keeper,
	}
}

func (tx *Tx) AnalyzeStart(tag string) {
	analyzer.StartTxLog(tag)
}

func (tx *Tx) AnalyzeStop(tag string) {
	analyzer.StopTxLog(tag)
}

// SaveTx check Tx do not transition state db
func (tx *Tx) SaveTx(msg *types.MsgEthereumTx) {}

// GetSenderAccount check Tx do not need this
func (tx *Tx) GetSenderAccount() authexported.Account { return nil }

// ResetWatcher check Tx do not need this
func (tx *Tx) ResetWatcher(account authexported.Account) {}

// RefundFeesWatcher refund the watcher, check Tx do not save state so. skip
func (tx *Tx) RefundFeesWatcher(account authexported.Account, coins sdk.Coins, price *big.Int) {}

// RestoreWatcherTransactionReceipt check Tx do not need restore
func (tx *Tx) RestoreWatcherTransactionReceipt(msg *types.MsgEthereumTx) {}

// Commit check Tx do not need
func (tx *Tx) Commit(msg *types.MsgEthereumTx, result *Result) {}

// FinalizeWatcher check Tx do not need this
func (tx *Tx) FinalizeWatcher(account authexported.Account, err error) {}
