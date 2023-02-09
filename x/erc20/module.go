package erc20

import (
	"encoding/json"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/types/upgrade"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/params"
	"github.com/FiboChain/fbc/libs/ibc-go/modules/core/base"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/client/context"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/types/module"
	abci "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	"github.com/FiboChain/fbc/x/erc20/client/cli"
	"github.com/FiboChain/fbc/x/erc20/keeper"
	"github.com/FiboChain/fbc/x/erc20/types"
)

var _ module.AppModuleBasic = AppModuleBasic{}
var _ module.AppModule = AppModule{}
var _ upgrade.UpgradeModule = AppModule{}

// AppModuleBasic struct
type AppModuleBasic struct{}

// Name for app module basic
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterCodec registers types for module
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// DefaultGenesis is json default structure
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return nil
}

// ValidateGenesis is the validation check of the Genesis
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	return nil
}

// RegisterRESTRoutes Registers rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
}

// GetQueryCmd Gets the root query command of this module
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetQueryCmd(types.ModuleName, cdc)
}

// GetTxCmd Gets the root tx command of this module
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return nil
}

//____________________________________________________________________________

// AppModule implements an application module for the erc20 module.
type AppModule struct {
	*base.BaseIBCUpgradeModule
	AppModuleBasic
	keeper Keeper
}

// NewAppModule creates a new AppModule Object
func NewAppModule(k Keeper) AppModule {
	ret := AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
	}
	ret.BaseIBCUpgradeModule = base.NewBaseIBCUpgradeModule(ret)
	return ret
}

// Name is module name
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants interface for registering invariants
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// Route specifies path for transactions
func (am AppModule) Route() string {
	return types.RouterKey
}

// NewHandler sets up a new handler for module
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

// QuerierRoute sets up path for queries
func (am AppModule) QuerierRoute() string {
	return types.ModuleName
}

// NewQuerierHandler sets up new querier handler for module
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return keeper.NewQuerier(am.keeper)
}

// BeginBlock function for module at start of each block
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {}

// EndBlock function for module at end of block
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// InitGenesis instantiates the genesis state
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	return nil
}

func (am AppModule) initGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	return InitGenesis(ctx, am.keeper, genesisState)
}

// ExportGenesis exports the genesis state to be used by daemon
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	return nil
}

func (am AppModule) RegisterTask() upgrade.HeightTask {
	return upgrade.NewHeightTask(0, func(ctx sdk.Context) error {
		if am.Sealed() {
			return nil
		}
		am.initGenesis(ctx, types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState()))
		return nil
	})
}

func (am AppModule) RegisterParam() params.ParamSet {
	return nil
}
