package app

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth"
	authante "github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/ante"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/bank"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/supply"
	"github.com/FiboChain/fbc/x/evm"
	evmtypes "github.com/FiboChain/fbc/x/evm/types"
)

// feeCollectorHandler set or get the value of feeCollectorAcc
func updateFeeCollectorHandler(bk bank.Keeper, sk supply.Keeper) sdk.UpdateFeeCollectorAccHandler {
	return func(ctx sdk.Context, balance sdk.Coins) error {
		return bk.SetCoins(ctx, sk.GetModuleAddress(auth.FeeCollectorName), balance)
	}
}

// evmTxFeeHandler get tx fee for evm tx
func evmTxFeeHandler() sdk.GetTxFeeHandler {
	return func(ctx sdk.Context, tx sdk.Tx) (fee sdk.Coins, isEvm bool) {
		if evmTx, ok := tx.(*evmtypes.MsgEthereumTx); ok {
			isEvm = true
			_ = evmTx.VerifySig(evmTx.ChainID(), ctx.BlockHeight())

		}
		if feeTx, ok := tx.(authante.FeeTx); ok {
			fee = feeTx.GetFee()
		}

		return
	}
}

// fixLogForParallelTxHandler fix log for parallel tx
func fixLogForParallelTxHandler(ek *evm.Keeper) sdk.LogFix {
	return func(execResults [][]string) (logs [][]byte) {
		return ek.FixLog(execResults)
	}
}
