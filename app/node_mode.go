package app

import (
	"fmt"
	"sort"
	"strings"

	appconfig "github.com/FiboChain/fbc/app/config"
	"github.com/FiboChain/fbc/app/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/client/flags"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/server"
	store "github.com/FiboChain/fbc/libs/cosmos-sdk/store/iavl"
	"github.com/FiboChain/fbc/libs/iavl"
	abcitypes "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	"github.com/FiboChain/fbc/libs/tendermint/libs/log"
	"github.com/FiboChain/fbc/libs/tendermint/mempool"
	evmtypes "github.com/FiboChain/fbc/x/evm/types"
	"github.com/FiboChain/fbc/x/evm/watcher"
	"github.com/spf13/viper"
)

func setNodeConfig(ctx *server.Context) {
	nodeMode := viper.GetString(types.FlagNodeMode)

	ctx.Logger.Info("Starting node", "mode", nodeMode)

	switch types.NodeMode(nodeMode) {
	case types.RpcNode:
		setRpcConfig(ctx)
	case types.ValidatorNode:
		setValidatorConfig(ctx)
	case types.ArchiveNode:
		setArchiveConfig(ctx)
	default:
		if len(nodeMode) > 0 {
			ctx.Logger.Error(
				fmt.Sprintf("Wrong value (%s) is set for %s, the correct value should be one of %s, %s, and %s",
					nodeMode, types.FlagNodeMode, types.RpcNode, types.ValidatorNode, types.ArchiveNode))
		}
	}
}

func setRpcConfig(ctx *server.Context) {
	viper.SetDefault(abcitypes.FlagDisableABCIQueryMutex, true)
	viper.SetDefault(evmtypes.FlagEnableBloomFilter, true)
	viper.SetDefault(watcher.FlagFastQueryLru, 10000)
	viper.SetDefault(watcher.FlagFastQuery, true)
	viper.SetDefault(iavl.FlagIavlEnableAsyncCommit, true)
	viper.SetDefault(flags.FlagMaxOpenConnections, 20000)
	viper.SetDefault(mempool.FlagEnablePendingPool, true)
	viper.SetDefault(server.FlagCORS, "*")
	ctx.Logger.Info(fmt.Sprintf(
		"Set --%s=%v\n--%s=%v\n--%s=%v\n--%s=%v\n--%s=%v\n--%s=%v\n--%s=%v\n--%s=%v by rpc node mode",
		abcitypes.FlagDisableABCIQueryMutex, true, evmtypes.FlagEnableBloomFilter, true, watcher.FlagFastQueryLru, 10000,
		watcher.FlagFastQuery, true, iavl.FlagIavlEnableAsyncCommit, true,
		flags.FlagMaxOpenConnections, 20000, mempool.FlagEnablePendingPool, true,
		server.FlagCORS, "*"))
}

func setValidatorConfig(ctx *server.Context) {
	viper.SetDefault(abcitypes.FlagDisableABCIQueryMutex, true)
	viper.SetDefault(appconfig.FlagEnableDynamicGp, false)
	viper.SetDefault(iavl.FlagIavlEnableAsyncCommit, true)
	viper.SetDefault(store.FlagIavlCacheSize, 10000000)
	viper.SetDefault(server.FlagPruning, "everything")
	ctx.Logger.Info(fmt.Sprintf("Set --%s=%v\n--%s=%v\n--%s=%v\n--%s=%v\n--%s=%v by validator node mode",
		abcitypes.FlagDisableABCIQueryMutex, true, appconfig.FlagEnableDynamicGp, false, iavl.FlagIavlEnableAsyncCommit, true,
		store.FlagIavlCacheSize, 10000000, server.FlagPruning, "everything"))
}

func setArchiveConfig(ctx *server.Context) {
	viper.SetDefault(server.FlagPruning, "nothing")
	viper.SetDefault(abcitypes.FlagDisableABCIQueryMutex, true)
	viper.SetDefault(evmtypes.FlagEnableBloomFilter, true)
	viper.SetDefault(iavl.FlagIavlEnableAsyncCommit, true)
	viper.SetDefault(flags.FlagMaxOpenConnections, 20000)
	viper.SetDefault(server.FlagCORS, "*")
	ctx.Logger.Info(fmt.Sprintf(
		"Set --%s=%v\n--%s=%v\n--%s=%v\n--%s=%v\n--%s=%v\n--%s=%v by archive node mode",
		server.FlagPruning, "nothing", abcitypes.FlagDisableABCIQueryMutex, true, evmtypes.FlagEnableBloomFilter, true,
		iavl.FlagIavlEnableAsyncCommit, true, flags.FlagMaxOpenConnections, 20000,
		server.FlagCORS, "*"))
}

func logStartingFlags(logger log.Logger) {
	msg := "All flags:\n"

	var maxLen int
	kvMap := make(map[string]interface{})
	var keys []string
	for _, key := range viper.AllKeys() {

		if strings.Index(key, "stream.") == 0 {
			continue
		}
		if strings.Index(key, "backend.") == 0 {
			continue
		}

		keys = append(keys, key)
		kvMap[key] = viper.Get(key)
		if len(key) > maxLen {
			maxLen = len(key)
		}
	}

	sort.Strings(keys)
	for _, k := range keys {
		msg += fmt.Sprintf("	%-45s= %v\n", k, kvMap[k])
	}

	logger.Info(msg)
}
