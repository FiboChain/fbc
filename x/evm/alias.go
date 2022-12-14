package evm

import (
	"github.com/FiboChain/fbc/x/evm/keeper"
	"github.com/FiboChain/fbc/x/evm/types"
)

// nolint
const (
	ModuleName        = types.ModuleName
	StoreKey          = types.StoreKey
	RouterKey         = types.RouterKey
	DefaultParamspace = types.DefaultParamspace
)

// nolint
var (
	NewKeeper              = keeper.NewKeeper
	TxDecoder              = types.TxDecoder
	NewSimulateKeeper      = keeper.NewSimulateKeeper
	SetEvmParamsNeedUpdate = types.SetEvmParamsNeedUpdate
)

//nolint
type (
	Keeper       = keeper.Keeper
	GenesisState = types.GenesisState
)
