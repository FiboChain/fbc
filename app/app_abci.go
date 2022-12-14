package app

import (
	appconfig "github.com/FiboChain/fbc/app/config"
	abci "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	"github.com/FiboChain/fbc/libs/tendermint/trace"
	"github.com/FiboChain/fbc/x/common/analyzer"
	"github.com/FiboChain/fbc/x/evm"
)

// BeginBlock implements the Application interface
func (app *FBchainApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {

	analyzer.OnAppBeginBlockEnter(app.LastBlockHeight() + 1)

	// dump app.LastBlockHeight()-1 info for reactor sync mode
	trace.GetElapsedInfo().Dump(app.Logger())
	return app.BaseApp.BeginBlock(req)
}

func (app *FBchainApp) DeliverTx(req abci.RequestDeliverTx) (res abci.ResponseDeliverTx) {

	analyzer.OnAppDeliverTxEnter()

	resp := app.BaseApp.DeliverTx(req)

	if appconfig.GetOecConfig().GetEnableDynamicGp() {
		tx, err := evm.TxDecoder(app.Codec())(req.Tx)
		if err == nil {
			//optimize get tx gas price can not get value from verifySign method
			app.blockGasPrice = append(app.blockGasPrice, tx.GetGasPrice())
		}
	}

	return resp
}

// EndBlock implements the Application interface
func (app *FBchainApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {

	return app.BaseApp.EndBlock(req)
}

// Commit implements the Application interface
func (app *FBchainApp) Commit(req abci.RequestCommit) abci.ResponseCommit {

	defer analyzer.OnCommitDone()
	res := app.BaseApp.Commit(req)

	// we call watch#Commit here ,because
	// 1. this round commit a valid block
	// 2. before commit the block,State#updateToState hasent not called yet,so the proposalBlockPart is not nil which means we wont
	// 	  call the prerun during commit step(edge case)
	app.EvmKeeper.Watcher.Commit()

	return res
}
