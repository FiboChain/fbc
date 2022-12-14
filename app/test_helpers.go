package app

import (
	abci "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	"github.com/FiboChain/fbc/libs/tendermint/libs/log"
	dbm "github.com/FiboChain/fbc/libs/tm-db"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
)

// Setup initializes a new FBchainApp. A Nop logger is set in FBchainApp.
func Setup(isCheckTx bool) *FBchainApp {
	db := dbm.NewMemDB()
	app := NewFBchainApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, 0)

	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		genesisState := NewDefaultGenesisState()
		stateBytes, err := codec.MarshalJSONIndent(app.Codec(), genesisState)
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		app.InitChain(
			abci.RequestInitChain{
				Validators:    []abci.ValidatorUpdate{},
				AppStateBytes: stateBytes,
			},
		)
	}

	return app
}
