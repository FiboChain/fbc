package farm

import (
	"github.com/FiboChain/fbc/x/farm/keeper"
	"github.com/FiboChain/fbc/x/farm/types"
)

const (
	StoreKey            = types.StoreKey
	DefaultParamspace   = types.DefaultParamspace
	DefaultCodespace    = types.DefaultCodespace
	ModuleName          = types.ModuleName
	MintFarmingAccount  = types.MintFarmingAccount
	YieldFarmingAccount = types.YieldFarmingAccount
	RouterKey           = types.RouterKey
)

var (
	NewKeeper          = keeper.NewKeeper
	RegisterInvariants = keeper.RegisterInvariants
)

type (
	Keeper = keeper.Keeper
)
