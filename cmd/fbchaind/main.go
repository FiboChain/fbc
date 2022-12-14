package main

import (
	"encoding/json"
	"fmt"
	"github.com/FiboChain/fbc/app/logevents"
	"io"

	"github.com/FiboChain/fbc/app/rpc"
	evmtypes "github.com/FiboChain/fbc/x/evm/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	tmamino "github.com/FiboChain/fbc/libs/tendermint/crypto/encoding/amino"
	"github.com/FiboChain/fbc/libs/tendermint/crypto/multisig"
	"github.com/FiboChain/fbc/libs/tendermint/libs/cli"
	"github.com/FiboChain/fbc/libs/tendermint/libs/log"
	tmtypes "github.com/FiboChain/fbc/libs/tendermint/types"
	dbm "github.com/FiboChain/fbc/libs/tm-db"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/baseapp"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/client/flags"
	clientkeys "github.com/FiboChain/fbc/libs/cosmos-sdk/client/keys"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/crypto/keys"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/server"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth"

	"github.com/FiboChain/fbc/app"
	"github.com/FiboChain/fbc/app/codec"
	"github.com/FiboChain/fbc/app/crypto/ethsecp256k1"
	fbchain "github.com/FiboChain/fbc/app/types"
	"github.com/FiboChain/fbc/cmd/client"
	"github.com/FiboChain/fbc/x/genutil"
	genutilcli "github.com/FiboChain/fbc/x/genutil/client/cli"
	genutiltypes "github.com/FiboChain/fbc/x/genutil/types"
	"github.com/FiboChain/fbc/x/staking"
)

const flagInvCheckPeriod = "inv-check-period"

var invCheckPeriod uint

func main() {
	cobra.EnableCommandSorting = false

	cdc := codec.MakeCodec(app.ModuleBasics)

	tmamino.RegisterKeyType(ethsecp256k1.PubKey{}, ethsecp256k1.PubKeyName)
	tmamino.RegisterKeyType(ethsecp256k1.PrivKey{}, ethsecp256k1.PrivKeyName)
	multisig.RegisterKeyType(ethsecp256k1.PubKey{}, ethsecp256k1.PubKeyName)

	keys.CryptoCdc = cdc
	genutil.ModuleCdc = cdc
	genutiltypes.ModuleCdc = cdc
	clientkeys.KeysCdc = cdc

	config := sdk.GetConfig()
	fbchain.SetBech32Prefixes(config)
	fbchain.SetBip44CoinType(config)
	config.Seal()

	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "fbchaind",
		Short:             "FBChain App Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}
	// CLI commands to initialize the chain
	rootCmd.AddCommand(
		client.ValidateChainID(
			genutilcli.InitCmd(ctx, cdc, app.ModuleBasics, app.DefaultNodeHome),
		),
		genutilcli.CollectGenTxsCmd(ctx, cdc, auth.GenesisAccountIterator{}, app.DefaultNodeHome),
		genutilcli.MigrateGenesisCmd(ctx, cdc),
		genutilcli.GenTxCmd(
			ctx, cdc, app.ModuleBasics, staking.AppModuleBasic{}, auth.GenesisAccountIterator{},
			app.DefaultNodeHome, app.DefaultCLIHome,
		),
		genutilcli.ValidateGenesisCmd(ctx, cdc, app.ModuleBasics),
		client.TestnetCmd(ctx, cdc, app.ModuleBasics, auth.GenesisAccountIterator{}),
		replayCmd(ctx, client.RegisterAppFlag),
		repairStateCmd(ctx),
		// AddGenesisAccountCmd allows users to add accounts to the genesis file
		AddGenesisAccountCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome),
		flags.NewCompletionCmd(rootCmd, true),
		dataCmd(ctx),
		exportAppCmd(ctx),
		iaviewerCmd(ctx, cdc),
		subscribeCmd(cdc),
	)

	subFunc := func(logger log.Logger) log.Subscriber {
		return logevents.NewProvider(logger)
	}
	// Tendermint node base commands
	server.AddCommands(ctx, cdc, rootCmd, newApp, closeApp, exportAppStateAndTMValidators,
		registerRoutes, client.RegisterAppFlag, app.PreRun, subFunc)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "FBCHAIN", app.DefaultNodeHome)
	rootCmd.PersistentFlags().UintVar(&invCheckPeriod, flagInvCheckPeriod,
		0, "Assert registered invariants every N blocks")
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func closeApp(iApp abci.Application) {
	fmt.Println("Close App")
	app := iApp.(*app.FBchainApp)
	app.StopStore()
	evmtypes.CloseIndexer()
	evmtypes.CloseTracer()
	rpc.CloseEthBackend()
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	pruningOpts, err := server.GetPruningOptionsFromFlags()
	if err != nil {
		panic(err)
	}

	return app.NewFBchainApp(
		logger,
		db,
		traceStore,
		true,
		map[int64]bool{},
		0,
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
		baseapp.SetHaltHeight(uint64(viper.GetInt(server.FlagHaltHeight))),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	var ethermintApp *app.FBchainApp
	if height != -1 {
		ethermintApp = app.NewFBchainApp(logger, db, traceStore, false, map[int64]bool{}, 0)

		if err := ethermintApp.LoadHeight(height); err != nil {
			return nil, nil, err
		}
	} else {
		ethermintApp = app.NewFBchainApp(logger, db, traceStore, true, map[int64]bool{}, 0)
	}

	return ethermintApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}
