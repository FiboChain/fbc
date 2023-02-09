package app

import (
	"runtime"
	"time"

	appconfig "github.com/FiboChain/fbc/app/config"
	"github.com/FiboChain/fbc/libs/system/trace"
	abci "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	"github.com/FiboChain/fbc/x/wasm/watcher"
)

// BeginBlock implements the Application interface
func (app *FBChainApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	trace.OnAppBeginBlockEnter(app.LastBlockHeight() + 1)
	app.EvmKeeper.Watcher.DelayEraseKey()
	return app.BaseApp.BeginBlock(req)
}

func (app *FBChainApp) DeliverTx(req abci.RequestDeliverTx) (res abci.ResponseDeliverTx) {

	trace.OnAppDeliverTxEnter()

	resp := app.BaseApp.DeliverTx(req)

	return resp
}

func (app *FBChainApp) PreDeliverRealTx(req []byte) (res abci.TxEssentials) {
	return app.BaseApp.PreDeliverRealTx(req)
}

func (app *FBChainApp) DeliverRealTx(req abci.TxEssentials) (res abci.ResponseDeliverTx) {
	trace.OnAppDeliverTxEnter()
	resp := app.BaseApp.DeliverRealTx(req)
	app.EvmKeeper.Watcher.RecordTxAndFailedReceipt(req, &resp, app.GetTxDecoder())

	return resp
}

// EndBlock implements the Application interface
func (app *FBChainApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	return app.BaseApp.EndBlock(req)
}

// Commit implements the Application interface
func (app *FBChainApp) Commit(req abci.RequestCommit) abci.ResponseCommit {
	if gcInterval := appconfig.GetFecConfig().GetGcInterval(); gcInterval > 0 {
		if (app.BaseApp.LastBlockHeight()+1)%int64(gcInterval) == 0 {
			startTime := time.Now()
			runtime.GC()
			elapsed := time.Now().Sub(startTime).Milliseconds()
			app.Logger().Info("force gc for debug", "height", app.BaseApp.LastBlockHeight()+1,
				"elapsed(ms)", elapsed)
		}
	}
	//defer trace.GetTraceSummary().Dump()
	defer trace.OnCommitDone()

	tasks := app.heightTasks[app.BaseApp.LastBlockHeight()+1]
	if tasks != nil {
		ctx := app.BaseApp.GetDeliverStateCtx()
		for _, t := range *tasks {
			if err := t.Execute(ctx); nil != err {
				panic("bad things")
			}
		}
	}
	res := app.BaseApp.Commit(req)

	// we call watch#Commit here ,because
	// 1. this round commit a valid block
	// 2. before commit the block,State#updateToState hasent not called yet,so the proposalBlockPart is not nil which means we wont
	// 	  call the prerun during commit step(edge case)
	app.EvmKeeper.Watcher.Commit()
	watcher.Commit()

	return res
}
