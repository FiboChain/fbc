package server

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	cmn "github.com/FiboChain/fbc/libs/tendermint/libs/os"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/baseapp"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/client/flags"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/store/flatkv"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/store/iavl"
	"github.com/FiboChain/fbc/libs/system"
	"github.com/FiboChain/fbc/libs/tendermint/state"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	storetypes "github.com/FiboChain/fbc/libs/cosmos-sdk/store/types"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	tmiavl "github.com/FiboChain/fbc/libs/iavl"
	abci "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	tmtypes "github.com/FiboChain/fbc/libs/tendermint/types"
)

// exchain full-node start flags
const (
	FlagListenAddr         = "rest.laddr"
	FlagUlockKey           = "rest.unlock_key"
	FlagUlockKeyHome       = "rest.unlock_key_home"
	FlagRestPathPrefix     = "rest.path_prefix"
	FlagCORS               = "cors"
	FlagMaxOpenConnections = "max-open"
	FlagHookstartInProcess = "startInProcess"
	FlagWebsocket          = "wsport"
	FlagWsMaxConnections   = "ws.max_connections"
	FlagWsSubChannelLength = "ws.sub_channel_length"
)

//module hook

type fnHookstartInProcess func(ctx *Context) error

type serverHookTable struct {
	hookTable map[string]interface{}
}

var gSrvHookTable = serverHookTable{make(map[string]interface{})}

func InstallHookEx(flag string, hooker fnHookstartInProcess) {
	gSrvHookTable.hookTable[flag] = hooker
}

//call hooker function
func callHooker(flag string, args ...interface{}) error {
	params := make([]interface{}, 0)
	switch flag {
	case FlagHookstartInProcess:
		{
			//none hook func, return nil
			function, ok := gSrvHookTable.hookTable[FlagHookstartInProcess]
			if !ok {
				return nil
			}
			params = append(params, args...)
			if len(params) != 1 {
				return errors.New("too many or less parameter called, want 1")
			}

			//param type check
			p1, ok := params[0].(*Context)
			if !ok {
				return errors.New("wrong param 1 type. want *Context, got" + reflect.TypeOf(params[0]).String())
			}

			//get hook function and call it
			caller := function.(fnHookstartInProcess)
			return caller(p1)
		}
	default:
		break
	}
	return nil
}

//end of hook

func setPID(ctx *Context) {
	pid := os.Getpid()
	f, err := os.OpenFile(filepath.Join(ctx.Config.RootDir, "config", "pid"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		cmn.Exit(err.Error())
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	_, err = writer.WriteString(strconv.Itoa(pid))
	if err != nil {
		fmt.Println(err.Error())
	}
	writer.Flush()
}

// StopCmd stop the node gracefully
// Tendermint.
func StopCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the node gracefully",
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(filepath.Join(ctx.Config.RootDir, "config", "pid"))
			if err != nil {
				errStr := fmt.Sprintf("%s Please finish the process of fbchaind through kill -2 pid to stop gracefully", err.Error())
				cmn.Exit(errStr)
			}
			defer f.Close()
			in := bufio.NewScanner(f)
			in.Scan()
			pid, err := strconv.Atoi(in.Text())
			if err != nil {
				errStr := fmt.Sprintf("%s Please finish the process of fbchaind through kill -2 pid to stop gracefully", err.Error())
				cmn.Exit(errStr)
			}
			process, err := os.FindProcess(pid)
			if err != nil {
				cmn.Exit(err.Error())
			}
			err = process.Signal(os.Interrupt)
			if err != nil {
				cmn.Exit(err.Error())
			}
			fmt.Println("pid", pid, "has been sent SIGINT")
			return nil
		},
	}
	return cmd
}

var sem *nodeSemaphore

type nodeSemaphore struct {
	done chan struct{}
}

func Stop() {
	sem.done <- struct{}{}
}

// RegisterServerFlags registers the flags required for rest server
func RegisterServerFlags(cmd *cobra.Command) *cobra.Command {
	// core flags for the ABCI application
	cmd.Flags().String(flagAddress, "tcp://0.0.0.0:26658", "Listen address")
	cmd.Flags().String(flagTraceStore, "", "Enable KVStore tracing to an output file")
	cmd.Flags().Bool(FlagTrace, false, "Provide full stack traces for errors in ABCI Log")
	cmd.Flags().String(
		FlagMinGasPrices, "",
		"Minimum gas prices to accept for transactions; Any fee in a tx must meet this minimum (e.g. 0.01photino;0.0001stake)",
	)
	cmd.Flags().IntSlice(FlagUnsafeSkipUpgrades, []int{}, "Skip a set of upgrade heights to continue the old binary")
	cmd.Flags().Uint64(FlagHaltHeight, 0, "Block height at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Uint64(FlagHaltTime, 0, "Minimum block time (in Unix seconds) at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Bool(FlagInterBlockCache, true, "Enable inter-block caching")
	cmd.Flags().String(flagCPUProfile, "", "Enable CPU profiling and write to the provided file")

	cmd.Flags().String(FlagPruning, storetypes.PruningOptionDefault, "Pruning strategy (default|nothing|everything|custom)")
	cmd.Flags().Uint64(FlagPruningKeepRecent, 0, "Number of recent heights to keep on disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(FlagPruningKeepEvery, 0, "Offset heights to keep on disk after 'keep-every' (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(FlagPruningInterval, 0, "Height interval at which pruned heights are removed from disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(FlagPruningMaxWsNum, 0, "Max number of historic states to keep on disk (ignored if pruning is not 'custom')")
	cmd.Flags().String(FlagLocalRpcPort, "", "Local rpc port for mempool and block monitor on cosmos layer(ignored if mempool/block monitoring is not required)")
	cmd.Flags().String(FlagPortMonitor, "", "Local target ports for connecting number monitoring(ignored if connecting number monitoring is not required)")
	cmd.Flags().String(FlagEvmImportMode, "default", "Select import mode for evm state (default|files|db)")
	cmd.Flags().String(FlagEvmImportPath, "", "Evm contract & storage db or files used for InitGenesis")
	cmd.Flags().Uint64(FlagGoroutineNum, 0, "Limit on the number of goroutines used to import evm data(ignored if evm-import-mode is 'default')")

	cmd.Flags().Bool(tmtypes.FlagDownloadDDS, false, "Download delta")
	cmd.Flags().Bool(tmtypes.FlagUploadDDS, false, "Upload delta")
	cmd.Flags().Bool(tmtypes.FlagAppendPid, false, "Append pid to the identity of delta producer")
	cmd.Flags().String(tmtypes.FlagRedisUrl, "localhost:6379", "redis url")
	cmd.Flags().String(tmtypes.FlagRedisAuth, "", "redis auth")
	cmd.Flags().Int(tmtypes.FlagRedisExpire, 300, "delta expiration time. unit is second")
	cmd.Flags().Int(tmtypes.FlagRedisDB, 0, "delta db num")
	cmd.Flags().Int(tmtypes.FlagDDSCompressType, 0, "delta compress type. 0|1|2|3")
	cmd.Flags().Int(tmtypes.FlagDDSCompressFlag, 0, "delta compress flag. 0|1|2")
	cmd.Flags().Int(tmtypes.FlagBufferSize, 10, "delta buffer size")
	cmd.Flags().String(FlagLogServerUrl, "", "log server url")
	cmd.Flags().Int(tmtypes.FlagDeltaVersion, tmtypes.DeltaVersion, "Specify delta version")

	cmd.Flags().Int(iavl.FlagIavlCacheSize, 1000000, "Max size of iavl cache")
	cmd.Flags().Float64(tmiavl.FlagIavlCacheInitRatio, 0, "iavl cache init ratio, 0.0~1.0, default is 0, iavl cache map would be init with (cache size * init ratio)")
	cmd.Flags().StringToInt(tmiavl.FlagOutputModules, map[string]int{"evm": 1, "acc": 1}, "decide which module in iavl to be printed")
	cmd.Flags().Int64(tmiavl.FlagIavlCommitIntervalHeight, 100, "Max interval to commit node cache into leveldb")
	cmd.Flags().Int64(tmiavl.FlagIavlMinCommitItemCount, 500000, "Min nodes num to triggle node cache commit") //触发节点缓存提交的最小节点数
	cmd.Flags().Int(tmiavl.FlagIavlHeightOrphansCacheSize, 8, "Max orphan version to cache in memory")         //缓存在内存中的最大孤儿版本
	cmd.Flags().Int(tmiavl.FlagIavlMaxCommittedHeightNum, 30, "Max committed version to cache in memory")      //缓存在内存中的最大提交版本
	cmd.Flags().Bool(tmiavl.FlagIavlEnableAsyncCommit, false, "Enable async commit")
	cmd.Flags().Bool(abci.FlagDisableABCIQueryMutex, false, "Disable local client query mutex for better concurrency")
	cmd.Flags().Bool(abci.FlagDisableCheckTx, false, "Disable checkTx for test")
	cmd.Flags().MarkHidden(abci.FlagDisableCheckTx)
	cmd.Flags().Bool(abci.FlagCloseMutex, false, fmt.Sprintf("Deprecated in v0.19.13 version, use --%s instead.", abci.FlagDisableABCIQueryMutex))
	cmd.Flags().MarkHidden(abci.FlagCloseMutex)
	cmd.Flags().Bool(FlagExportKeystore, false, "export keystore file when call newaccount ")
	cmd.Flags().Bool(system.FlagEnableGid, false, "Display goroutine id in log")

	cmd.Flags().Int(state.FlagApplyBlockPprofTime, -1, "time(ms) of executing ApplyBlock, if it is higher than this value, save pprof")

	cmd.Flags().Float64Var(&baseapp.GasUsedFactor, baseapp.FlagGasUsedFactor, 0.4, "factor to calculate history gas used")

	cmd.Flags().Bool(sdk.FlagMultiCache, false, "Enable multi cache")
	cmd.Flags().Int(sdk.MaxAccInMultiCache, 0, "max acc in multi cache")
	cmd.Flags().Int(sdk.MaxStorageInMultiCache, 0, "max storage in multi cache")
	cmd.Flags().Bool(flatkv.FlagEnable, false, "Enable flat kv storage for read performance")

	// Don`t use cmd.Flags().*Var functions(such as cmd.Flags.IntVar) here, because it doesn't work with environment variables.
	// Use setExternalPackageValue function instead.
	viper.BindPFlag(FlagTrace, cmd.Flags().Lookup(FlagTrace))
	viper.BindPFlag(FlagPruning, cmd.Flags().Lookup(FlagPruning))
	viper.BindPFlag(FlagPruningKeepRecent, cmd.Flags().Lookup(FlagPruningKeepRecent))
	viper.BindPFlag(FlagPruningKeepEvery, cmd.Flags().Lookup(FlagPruningKeepEvery))
	viper.BindPFlag(FlagPruningInterval, cmd.Flags().Lookup(FlagPruningInterval))
	viper.BindPFlag(FlagPruningMaxWsNum, cmd.Flags().Lookup(FlagPruningMaxWsNum))
	viper.BindPFlag(FlagLocalRpcPort, cmd.Flags().Lookup(FlagLocalRpcPort))
	viper.BindPFlag(FlagPortMonitor, cmd.Flags().Lookup(FlagPortMonitor))
	viper.BindPFlag(FlagEvmImportMode, cmd.Flags().Lookup(FlagEvmImportMode))
	viper.BindPFlag(FlagEvmImportPath, cmd.Flags().Lookup(FlagEvmImportPath))
	viper.BindPFlag(FlagGoroutineNum, cmd.Flags().Lookup(FlagGoroutineNum))

	cmd.Flags().Bool(state.FlagParalleledTx, false, "Enable Parallel Tx")

	cmd.Flags().String(FlagListenAddr, "tcp://0.0.0.0:26659", "EVM RPC and cosmos-sdk REST API listen address.")
	cmd.Flags().String(FlagUlockKey, "", "Select the keys to unlock on the RPC server")
	cmd.Flags().String(FlagUlockKeyHome, os.ExpandEnv("$HOME/.fbchaincli"), "The keybase home path")
	cmd.Flags().String(FlagRestPathPrefix, "fbchain", "Path prefix for registering rest api route.")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")
	cmd.Flags().String(FlagCORS, "", "Set the rest-server domains that can make CORS requests (* for all)")
	cmd.Flags().Int(FlagMaxOpenConnections, 1000, "The number of maximum open connections of rest-server")
	cmd.Flags().String(FlagWebsocket, "8546", "websocket port to listen to")
	cmd.Flags().Int(FlagWsMaxConnections, 20000, "the max capacity number of websocket client connections")
	cmd.Flags().Int(FlagWsSubChannelLength, 100, "the length of subscription channel")
	cmd.Flags().String(flags.FlagChainID, "", "Chain ID of tendermint node for web3")
	cmd.Flags().StringP(flags.FlagBroadcastMode, "b", flags.BroadcastSync, "Transaction broadcasting mode (sync|async|block) for web3")
	return cmd
}

func nodeModeCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node-mode",
		Short: "fbchaind start --node-mode help info",
		Long: `There are three node modes that can be set when the fbchaind start
set --node-mode=rpc to manage the following flags:
	--enable-bloom-filter=true
	--fast-lru=10000
	--fast-query=true
	--iavl-enable-async-commit=true
	--max-open=20000
	--mempool.enable_pending_pool=true
	--cors=*

set --node-mode=validator to manage the following flags:
	--disable-query-mutex=true
	--enable-dynamic-gp=false
	--iavl-enable-async-commit=true
	--iavl-cache-size=10000000
	--pruning=everything

set --node-mode=archive to manage the following flags:
	--pruning=nothing
	--disable-query-mutex=true
	--enable-bloom-filter=true
	--iavl-enable-async-commit=true
	--max-open=20000
	--cors=*`,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	return cmd
}
