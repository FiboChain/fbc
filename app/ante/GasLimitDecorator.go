package ante

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	sdkerrors "github.com/FiboChain/fbc/libs/cosmos-sdk/types/errors"
	evmtypes "github.com/FiboChain/fbc/x/evm/types"
)

// EVMKeeper defines the expected keeper interface used on the Eth AnteHandler
type EVMKeeper interface {
	GetParams(ctx sdk.Context) evmtypes.Params
	IsAddressBlocked(ctx sdk.Context, addr sdk.AccAddress) bool
}

// NewGasLimitDecorator creates a new GasLimitDecorator.
func NewGasLimitDecorator(evm EVMKeeper) GasLimitDecorator {
	return GasLimitDecorator{
		evm: evm,
	}
}

type GasLimitDecorator struct {
	evm EVMKeeper
}

// AnteHandle handles incrementing the sequence of the sender.
func (g GasLimitDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	pinAnte(ctx.AnteTracer(), "GasLimitDecorator")

	if tx.GetGas() > g.evm.GetParams(ctx).MaxGasLimitPerTx {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrTxTooLarge, "too large gas limit, it must be less than %d", g.evm.GetParams(ctx).MaxGasLimitPerTx)
	}

	return next(ctx, tx, simulate)
}
