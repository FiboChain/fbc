package keeper

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	evmtypes "github.com/FiboChain/fbc/x/evm/types"
)

type EvmKeeper interface {
	GetChainConfig(ctx sdk.Context) (evmtypes.ChainConfig, bool)
	GenerateCSDBParams() evmtypes.CommitStateDBParams
	GetParams(ctx sdk.Context) evmtypes.Params
	AddInnerTx(...interface{})
	AddContract(...interface{})
}
