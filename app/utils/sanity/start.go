package sanity

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/FiboChain/fbc/app/config"
	apptype "github.com/FiboChain/fbc/app/types"
	"github.com/FiboChain/fbc/app/utils/appstatus"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/server"
	cosmost "github.com/FiboChain/fbc/libs/cosmos-sdk/store/types"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/iavl"
	"github.com/FiboChain/fbc/libs/tendermint/consensus"
	"github.com/FiboChain/fbc/libs/tendermint/state"
	"github.com/FiboChain/fbc/libs/tendermint/types"
	"github.com/FiboChain/fbc/x/evm/watcher"
	"github.com/FiboChain/fbc/x/infura"
)

// CheckStart check start command's flags. if user set conflict flags return error.
// the conflicts flags are:
// --fast-query      conflict with --pruning=nothing
// --enable-preruntx conflict with --download-delta
//
// based the conflicts above and node-mode below
// --node-mode=rpc manage the following flags:
//     --disable-checktx-mutex=true
//     --disable-query-mutex=true
//     --enable-bloom-filter=true
//     --fast-lru=10000
//     --fast-query=true
//     --iavl-enable-async-commit=true
//     --max-open=20000
//     --mempool.enable_pending_pool=true
//     --cors=*
//
// --node-mode=validator manage the following flags:
//     --disable-checktx-mutex=true
//     --disable-query-mutex=true
//     --dynamic-gp-mode=2
//     --iavl-enable-async-commit=true
//     --iavl-cache-size=10000000
//     --pruning=everything
//
// --node-mode=archive manage the following flags:
//    --pruning=nothing
//    --disable-checktx-mutex=true
//    --disable-query-mutex=true
//    --enable-bloom-filter=true
//    --iavl-enable-async-commit=true
//    --max-open=20000
//    --cors=*
//
// then
// --node-mode=archive(--pruning=nothing) conflicts with --fast-query

var (
	startDependentElems = []dependentPair{
		{ // if infura.FlagEnable=true , watcher.FlagFastQuery must be set to true
			config:       boolItem{name: infura.FlagEnable, expect: true},
			reliedConfig: boolItem{name: watcher.FlagFastQuery, expect: true},
		},
	}
	// conflicts flags
	startConflictElems = []conflictPair{
		// --fast-query      conflict with --pruning=nothing
		{
			configA: boolItem{name: watcher.FlagFastQuery, expect: true},
			configB: stringItem{name: server.FlagPruning, expect: cosmost.PruningOptionNothing},
		},
		// --enable-preruntx conflict with --download-delta
		{
			configA: boolItem{name: consensus.EnablePrerunTx, expect: true},
			configB: boolItem{name: types.FlagDownloadDDS, expect: true},
		},
		// --multi-cache conflict with --download-delta
		{
			configA: boolItem{name: sdk.FlagMultiCache, expect: true},
			configB: boolItem{name: types.FlagDownloadDDS, expect: true},
		},
		{
			configA: stringItem{name: apptype.FlagNodeMode, expect: string(apptype.RpcNode)},
			configB: stringItem{name: server.FlagPruning, expect: cosmost.PruningOptionNothing},
		},
		// --node-mode=archive(--pruning=nothing) conflicts with --fast-query
		{
			configA: stringItem{name: apptype.FlagNodeMode, expect: string(apptype.ArchiveNode)},
			configB: boolItem{name: watcher.FlagFastQuery, expect: true},
		},
		{
			configA: stringItem{name: apptype.FlagNodeMode, expect: string(apptype.RpcNode)},
			configB: boolItem{name: config.FlagEnablePGU, expect: true},
		},
		{
			configA: stringItem{name: apptype.FlagNodeMode, expect: string(apptype.ArchiveNode)},
			configB: boolItem{name: config.FlagEnablePGU, expect: true},
		},
		{
			configA: stringItem{name: apptype.FlagNodeMode, expect: string(apptype.InnertxNode)},
			configB: boolItem{name: config.FlagEnablePGU, expect: true},
		},
		{
			configA: boolItem{name: iavl.FlagIavlEnableFastStorage, expect: true},
			configB: funcItem{name: "Upgraded to fast IAVL", expect: false, f: appstatus.IsFastStorageStrategy},
			tips: fmt.Sprintf("Upgrade to IAVL fast storage may take several hours, "+
				"you can use fbchaind fss create command to upgrade, or unset --%v", iavl.FlagIavlEnableFastStorage),
		},
	}

	checkRangeItems = []rangeItem{
		{
			enumRange: []int{int(state.DeliverTxsExecModeSerial), state.DeliverTxsExecModeParallel},
			name:      state.FlagDeliverTxsExecMode,
		},
	}
)

// CheckStart check start command.If it has conflict pair above. then return the conflict error
func CheckStart() error {
	if viper.GetBool(FlagDisableSanity) {
		return nil
	}
	for _, v := range startDependentElems {
		if err := v.check(); err != nil {
			return err
		}
	}
	for _, v := range startConflictElems {
		if err := v.check(); err != nil {
			return err
		}
	}

	for _, v := range checkRangeItems {
		if err := v.checkRange(); err != nil {
			return err
		}
	}

	return nil
}
