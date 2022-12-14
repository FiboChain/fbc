package keeper

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/supply"
	"github.com/FiboChain/fbc/libs/tendermint/libs/log"
	swap "github.com/FiboChain/fbc/x/ammswap/keeper"
	evm "github.com/FiboChain/fbc/x/evm/keeper"
	"github.com/FiboChain/fbc/x/farm/types"
	"github.com/FiboChain/fbc/x/token"
)

// Keeper of the farm store
type Keeper struct {
	storeKey         sdk.StoreKey
	cdc              *codec.Codec
	feeCollectorName string // name of the FeeCollector ModuleAccount
	paramSubspace    types.ParamSubspace
	supplyKeeper     supply.Keeper
	tokenKeeper      token.Keeper
	swapKeeper       swap.Keeper
	evmKeeper        evm.Keeper
	govKeeper        GovKeeper
	ObserverKeeper   []types.BackendKeeper
}

// NewKeeper creates a farm keeper
func NewKeeper(feeCollectorName string, supplyKeeper supply.Keeper,
	tokenKeeper token.Keeper,
	swapKeeper swap.Keeper,
	evmKeeper evm.Keeper,
	paramSubspace types.ParamSubspace, key sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		storeKey:         key,
		cdc:              cdc,
		feeCollectorName: feeCollectorName,
		paramSubspace:    paramSubspace.WithKeyTable(types.ParamKeyTable()),
		supplyKeeper:     supplyKeeper,
		tokenKeeper:      tokenKeeper,
		swapKeeper:       swapKeeper,
		evmKeeper:        evmKeeper,
	}
}

func (k Keeper) StoreKey() sdk.StoreKey {
	return k.storeKey
}

func (k Keeper) SupplyKeeper() supply.Keeper {
	return k.supplyKeeper
}

func (k Keeper) TokenKeeper() token.Keeper {
	return k.tokenKeeper
}

func (k Keeper) SwapKeeper() swap.Keeper {
	return k.swapKeeper
}

func (k Keeper) EvmKeeper() evm.Keeper {
	return k.evmKeeper
}

// GetFeeCollector returns feeCollectorName
func (k Keeper) GetFeeCollector() string {
	return k.feeCollectorName
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

// SetGovKeeper sets keeper of gov
func (k *Keeper) SetGovKeeper(gk GovKeeper) {
	k.govKeeper = gk
}

func (k *Keeper) SetObserverKeeper(bk types.BackendKeeper) {
	k.ObserverKeeper = append(k.ObserverKeeper, bk)
}

func (k Keeper) OnClaim(ctx sdk.Context, address sdk.AccAddress, poolName string, claimedCoins sdk.SysCoins) {
	for _, observer := range k.ObserverKeeper {
		observer.OnFarmClaim(ctx, address, poolName, claimedCoins)
	}
}
