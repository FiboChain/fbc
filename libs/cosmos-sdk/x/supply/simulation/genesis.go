package simulation

// DONTCOVER

import (
	"fmt"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/types/module"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/supply/internal/types"
)

// RandomizedGenState generates a random GenesisState for supply
func RandomizedGenState(simState *module.SimulationState) {
	numAccs := int64(len(simState.Accounts))
	totalSupply := sdk.NewInt(simState.InitialStake * (numAccs + simState.NumBonded))
	supplyGenesis := types.NewGenesisState(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, totalSupply)))

	fmt.Printf("Generated supply parameters:\n%s\n", codec.MustMarshalJSONIndent(simState.Cdc, supplyGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(supplyGenesis)
}
