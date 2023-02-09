package app

import (
	"encoding/hex"
	"sort"
	"strings"

	appconfig "github.com/FiboChain/fbc/app/config"
	"github.com/FiboChain/fbc/app/gasprice"
	ethermint "github.com/FiboChain/fbc/app/types"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth"
	authante "github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/ante"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/bank"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/supply"
	abci "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	"github.com/FiboChain/fbc/libs/tendermint/types"
	"github.com/FiboChain/fbc/x/evm"
	evmtypes "github.com/FiboChain/fbc/x/evm/types"
)

// feeCollectorHandler set or get the value of feeCollectorAcc
func updateFeeCollectorHandler(bk bank.Keeper, sk supply.Keeper) sdk.UpdateFeeCollectorAccHandler {
	return func(ctx sdk.Context, balance sdk.Coins, txFeesplit []*sdk.FeeSplitInfo) error {
		err := bk.SetCoins(ctx, sk.GetModuleAccount(ctx, auth.FeeCollectorName).GetAddress(), balance)
		if err != nil {
			return err
		}

		// split fee
		// come from feesplit module
		if txFeesplit != nil {
			feesplits, sortAddrs := groupByAddrAndSortFeeSplits(txFeesplit)
			for _, addr := range sortAddrs {
				acc := sdk.MustAccAddressFromBech32(addr)
				err = sk.SendCoinsFromModuleToAccount(ctx, auth.FeeCollectorName, acc, feesplits[addr])
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// fixLogForParallelTxHandler fix log for parallel tx
func fixLogForParallelTxHandler(ek *evm.Keeper) sdk.LogFix {
	return func(tx []sdk.Tx, logIndex []int, hasEnterEvmTx []bool, anteErrs []error, resp []abci.ResponseDeliverTx) (logs [][]byte) {
		return ek.FixLog(tx, logIndex, hasEnterEvmTx, anteErrs, resp)
	}
}

func preDeliverTxHandler(ak auth.AccountKeeper) sdk.PreDeliverTxHandler {
	return func(ctx sdk.Context, tx sdk.Tx, onlyVerifySig bool) {
		if evmTx, ok := tx.(*evmtypes.MsgEthereumTx); ok {
			if evmTx.BaseTx.From == "" {
				_ = evmTxVerifySigHandler(ctx.ChainID(), ctx.BlockHeight(), evmTx)
			}

			if types.HigherThanMars(ctx.BlockHeight()) {
				return
			}

			if onlyVerifySig {
				return
			}

			if from := evmTx.AccountAddress(); from != nil {
				ak.LoadAccount(ctx, from)
			}
			if to := evmTx.Data.Recipient; to != nil {
				ak.LoadAccount(ctx, to.Bytes())
			}
		}
	}
}

func evmTxVerifySigHandler(chainID string, blockHeight int64, evmTx *evmtypes.MsgEthereumTx) error {
	chainIDEpoch, err := ethermint.ParseChainID(chainID)
	if err != nil {
		return err
	}
	err = evmTx.VerifySig(chainIDEpoch, blockHeight)
	if err != nil {
		return err
	}
	return nil
}

func getTxFeeHandler() sdk.GetTxFeeHandler {
	return func(tx sdk.Tx) (fee sdk.Coins) {
		if feeTx, ok := tx.(authante.FeeTx); ok {
			fee = feeTx.GetFee()
		}

		return
	}
}

// getTxFeeAndFromHandler get tx fee and from
func getTxFeeAndFromHandler(ak auth.AccountKeeper) sdk.GetTxFeeAndFromHandler {
	return func(ctx sdk.Context, tx sdk.Tx) (fee sdk.Coins, isEvm bool, from string, to string, err error) {
		if evmTx, ok := tx.(*evmtypes.MsgEthereumTx); ok {
			isEvm = true
			err = evmTxVerifySigHandler(ctx.ChainID(), ctx.BlockHeight(), evmTx)
			if err != nil {
				return
			}
			fee = evmTx.GetFee()
			from = evmTx.BaseTx.From
			if len(from) > 2 {
				from = strings.ToLower(from[2:])
			}
			if evmTx.To() != nil {
				to = strings.ToLower(evmTx.To().String()[2:])
			}
		} else if feeTx, ok := tx.(authante.FeeTx); ok {
			fee = feeTx.GetFee()
			feePayer := feeTx.FeePayer(ctx)
			feePayerAcc := ak.GetAccount(ctx, feePayer)
			from = hex.EncodeToString(feePayerAcc.GetAddress())
		}

		return
	}
}

// groupByAddrAndSortFeeSplits
// feesplits must be ordered, not map(random),
// to ensure that the account number of the withdrawer(new account) is consistent
func groupByAddrAndSortFeeSplits(txFeesplit []*sdk.FeeSplitInfo) (feesplits map[string]sdk.Coins, sortAddrs []string) {
	feesplits = make(map[string]sdk.Coins)
	for _, f := range txFeesplit {
		feesplits[f.Addr.String()] = feesplits[f.Addr.String()].Add(f.Fee...)
	}
	if len(feesplits) == 0 {
		return
	}

	sortAddrs = make([]string, len(feesplits))
	index := 0
	for key := range feesplits {
		sortAddrs[index] = key
		index++
	}
	sort.Strings(sortAddrs)

	return
}

func updateGPOHandler(gpo *gasprice.Oracle) sdk.UpdateGPOHandler {
	return func(dynamicGpInfos []sdk.DynamicGasInfo) {
		if appconfig.GetFecConfig().GetDynamicGpMode() != ethermint.MinimalGpMode {
			for _, dgi := range dynamicGpInfos {
				gpo.CurrentBlockGPs.Update(dgi.GetGP(), dgi.GetGU())
			}
		}
	}
}
