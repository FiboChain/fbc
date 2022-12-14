package ante

import (
	"bytes"
	"math/big"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/exported"

	"github.com/ethereum/go-ethereum/common"
	ethcore "github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/baseapp"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	sdkerrors "github.com/FiboChain/fbc/libs/cosmos-sdk/types/errors"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/types"
	evmtypes "github.com/FiboChain/fbc/x/evm/types"
)

type AccountAnteDecorator struct {
	ak        auth.AccountKeeper
	sk        types.SupplyKeeper
	evmKeeper EVMKeeper
}

// NewAccountVerificationDecorator creates a new AccountVerificationDecorator
func NewAccountAnteDecorator(ak auth.AccountKeeper, ek EVMKeeper, sk types.SupplyKeeper) AccountAnteDecorator {
	return AccountAnteDecorator{
		ak:        ak,
		sk:        sk,
		evmKeeper: ek,
	}
}

func accountVerification(ctx *sdk.Context, acc exported.Account, tx *evmtypes.MsgEthereumTx) error {
	if ctx.BlockHeight() == 0 && acc.GetAccountNumber() != 0 {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidSequence,
			"invalid account number for height zero (got %d)", acc.GetAccountNumber(),
		)
	}

	evmDenom := sdk.DefaultBondDenom

	// validate sender has enough funds to pay for gas cost
	balance := acc.GetCoins().AmountOf(evmDenom)
	if balance.BigInt().Cmp(tx.Cost()) < 0 {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"sender balance < tx gas cost (%s%s < %s%s)", balance.String(), evmDenom, sdk.NewDecFromBigIntWithPrec(tx.Cost(), sdk.Precision).String(), evmDenom,
		)
	}
	return nil
}

func nonceVerificationInCheckTx(seq uint64, msgEthTx *evmtypes.MsgEthereumTx, isReCheckTx bool) error {
	if isReCheckTx {
		// recheckTx mode
		// sequence must strictly increasing
		if msgEthTx.Data.AccountNonce != seq {
			return sdkerrors.Wrapf(
				sdkerrors.ErrInvalidSequence,
				"invalid nonce; got %d, expected %d", msgEthTx.Data.AccountNonce, seq,
			)
		}
	} else {
		if baseapp.IsMempoolEnablePendingPool() {
			if msgEthTx.Data.AccountNonce < seq {
				return sdkerrors.Wrapf(
					sdkerrors.ErrInvalidSequence,
					"invalid nonce; got %d, expected %d", msgEthTx.Data.AccountNonce, seq,
				)
			}
		} else {
			// checkTx mode
			checkTxModeNonce := seq
			if !baseapp.IsMempoolEnableRecheck() {
				// if is enable recheck, the sequence of checkState will increase after commit(), so we do not need
				// to add pending txs len in the mempool.
				// but, if disable recheck, we will not increase sequence of checkState (even in force recheck case, we
				// will also reset checkState), so we will need to add pending txs len to get the right nonce
				gPool := baseapp.GetGlobalMempool()
				if gPool != nil {
					cnt := gPool.GetUserPendingTxsCnt(evmtypes.EthAddressStringer(common.BytesToAddress(msgEthTx.AccountAddress().Bytes())).String())
					checkTxModeNonce = seq + uint64(cnt)
				}
			}

			if baseapp.IsMempoolEnableSort() {
				if msgEthTx.Data.AccountNonce < seq || msgEthTx.Data.AccountNonce > checkTxModeNonce {
					return sdkerrors.Wrapf(
						sdkerrors.ErrInvalidSequence,
						"invalid nonce; got %d, expected in the range of [%d, %d]",
						msgEthTx.Data.AccountNonce, seq, checkTxModeNonce,
					)
				}
			} else {
				if msgEthTx.Data.AccountNonce != checkTxModeNonce {
					return sdkerrors.Wrapf(
						sdkerrors.ErrInvalidSequence,
						"invalid nonce; got %d, expected %d",
						msgEthTx.Data.AccountNonce, checkTxModeNonce,
					)
				}
			}
		}
	}
	return nil
}

func nonceVerification(ctx sdk.Context, acc exported.Account, msgEthTx *evmtypes.MsgEthereumTx) (sdk.Context, error) {
	seq := acc.GetSequence()
	// if multiple transactions are submitted in succession with increasing nonces,
	// all will be rejected except the first, since the first needs to be included in a block
	// before the sequence increments
	if ctx.IsCheckTx() {
		ctx.SetAccountNonce(seq)
		// will be checkTx and RecheckTx mode
		err := nonceVerificationInCheckTx(seq, msgEthTx, ctx.IsReCheckTx())
		if err != nil {
			return ctx, err
		}
	} else {
		// only deliverTx mode
		if msgEthTx.Data.AccountNonce != seq {
			return ctx, sdkerrors.Wrapf(
				sdkerrors.ErrInvalidSequence,
				"invalid nonce; got %d, expected %d", msgEthTx.Data.AccountNonce, seq,
			)
		}
	}
	return ctx, nil
}

func ethGasConsume(ctx *sdk.Context, acc exported.Account, accGetGas sdk.Gas, msgEthTx *evmtypes.MsgEthereumTx, simulate bool, sk types.SupplyKeeper) error {
	gasLimit := msgEthTx.GetGas()
	gas, err := ethcore.IntrinsicGas(msgEthTx.Data.Payload, []ethtypes.AccessTuple{}, msgEthTx.To() == nil, true, false)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to compute intrinsic gas cost")
	}

	// intrinsic gas verification during CheckTx
	if ctx.IsCheckTx() && gasLimit < gas {
		return sdkerrors.Wrapf(sdkerrors.ErrOutOfGas, "intrinsic gas too low: %d < %d", gasLimit, gas)
	}

	// Charge sender for gas up to limit
	if gasLimit != 0 {
		// Cost calculates the fees paid to validators based on gas limit and price
		cost := new(big.Int).Mul(msgEthTx.Data.Price, new(big.Int).SetUint64(gasLimit))

		evmDenom := sdk.DefaultBondDenom

		feeAmt := sdk.NewCoins(
			sdk.NewCoin(evmDenom, sdk.NewDecFromBigIntWithPrec(cost, sdk.Precision)), // int2dec
		)

		ctx.UpdateFromAccountCache(acc, accGetGas)

		err = auth.DeductFees(sk, *ctx, acc, feeAmt)
		if err != nil {
			return err
		}
	}

	// Set gas meter after ante handler to ignore gaskv costs
	auth.SetGasMeter(simulate, ctx, gasLimit)
	return nil
}

func incrementSeq(ctx sdk.Context, msgEthTx *evmtypes.MsgEthereumTx, ak auth.AccountKeeper, acc exported.Account) {
	if ctx.IsCheckTx() && !ctx.IsReCheckTx() && !baseapp.IsMempoolEnableRecheck() && !ctx.IsTraceTx() {
		return
	}

	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	ctx.SetGasMeter(sdk.NewInfiniteGasMeter())

	// increment sequence of all signers
	for _, addr := range msgEthTx.GetSigners() {
		var sacc exported.Account
		if acc != nil && bytes.Equal(addr, acc.GetAddress()) {
			// because we use infinite gas meter, we can don't care about the gas
			sacc = acc
		} else {
			sacc = ak.GetAccount(ctx, addr)
		}
		seq := sacc.GetSequence()
		if !baseapp.IsMempoolEnablePendingPool() {
			seq++
		} else if msgEthTx.Data.AccountNonce == seq {
			seq++
		}
		if err := sacc.SetSequence(seq); err != nil {
			panic(err)
		}
		ak.SetAccount(ctx, sacc)
	}
	return
}

func (avd AccountAnteDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgEthTx, ok := tx.(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", tx)
	}

	var acc exported.Account
	var getAccGasUsed sdk.Gas

	address := msgEthTx.AccountAddress()
	if address.Empty() && ctx.From() != "" {
		msgEthTx.SetFrom(ctx.From())
		address = msgEthTx.AccountAddress()
	}

	if !simulate {
		if address.Empty() {
			panic("sender address cannot be empty")
		}
		if ctx.IsCheckTx() {
			acc = avd.ak.GetAccount(ctx, address)
			if acc == nil {
				acc = avd.ak.NewAccountWithAddress(ctx, address)
				avd.ak.SetAccount(ctx, acc)
			}
			// on InitChain make sure account number == 0
			err = accountVerification(&ctx, acc, msgEthTx)
			if err != nil {
				return ctx, err
			}
		}

		acc, getAccGasUsed = getAccount(&avd.ak, &ctx, address, acc)
		if acc == nil {
			return ctx, sdkerrors.Wrapf(
				sdkerrors.ErrUnknownAddress,
				"account %s (%s) is nil", common.BytesToAddress(address.Bytes()), address,
			)
		}

		// account would not be updated
		ctx, err = nonceVerification(ctx, acc, msgEthTx)
		if err != nil {
			return ctx, err
		}

		// consume gas for compatible
		ctx.GasMeter().ConsumeGas(getAccGasUsed, "get account")

		ctx.EnableAccountCache()
		// account would be updated
		err = ethGasConsume(&ctx, acc, getAccGasUsed, msgEthTx, simulate, avd.sk)
		acc = nil
		acc, _ = ctx.GetFromAccountCacheData().(exported.Account)
		ctx.DisableAccountCache()
		if err != nil {
			return ctx, err
		}
	}

	incrementSeq(ctx, msgEthTx, avd.ak, acc)

	return next(ctx, tx, simulate)
}
