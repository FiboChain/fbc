package fee

import (
	"encoding/json"

	abci "github.com/FiboChain/fbc/libs/tendermint/abci/types"

	"github.com/FiboChain/fbc/libs/ibc-go/modules/apps/29-fee/types"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	fee "github.com/FiboChain/fbc/libs/ibc-go/modules/apps/29-fee"
	"github.com/FiboChain/fbc/libs/ibc-go/modules/apps/29-fee/keeper"
)

type TestFeeAppModuleBaisc struct {
	fee.AppModuleBasic
}

func (b TestFeeAppModuleBaisc) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState())
}

type TestFeeAppModule struct {
	fee.AppModule
	keeper keeper.Keeper
}

func NewTestFeeAppModule(keeper keeper.Keeper) *TestFeeAppModule {
	ret := &TestFeeAppModule{
		AppModule: fee.NewAppModule(keeper),
		keeper:    keeper,
	}
	return ret
}

func (a TestFeeAppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs := a.keeper.ExportGenesis(ctx)
	return types.ModuleCdc.MustMarshalJSON(gs)
}

func (a TestFeeAppModule) InitGenesis(ctx sdk.Context, message json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(message, &genesisState)
	a.keeper.InitGenesis(ctx, genesisState)
	return []abci.ValidatorUpdate{}
}
