package main

import (
	"fmt"
	authtypes "github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/types"

	"github.com/FiboChain/fbc/app"
	"github.com/FiboChain/fbc/app/codec"
	"github.com/FiboChain/fbc/app/crypto/ethsecp256k1"
	fbchain "github.com/FiboChain/fbc/app/types"
	"github.com/FiboChain/fbc/cmd/client"
	sdkclient "github.com/FiboChain/fbc/libs/cosmos-sdk/client"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/client/flags"
	clientkeys "github.com/FiboChain/fbc/libs/cosmos-sdk/client/keys"
	clientrpc "github.com/FiboChain/fbc/libs/cosmos-sdk/client/rpc"
	sdkcodec "github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/crypto/keys"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/version"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth"
	authcmd "github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/client/cli"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/client/utils"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/bank"
	tmamino "github.com/FiboChain/fbc/libs/tendermint/crypto/encoding/amino"
	"github.com/FiboChain/fbc/libs/tendermint/crypto/multisig"
	"github.com/FiboChain/fbc/libs/tendermint/libs/cli"
	"github.com/FiboChain/fbc/x/dex"
	evmtypes "github.com/FiboChain/fbc/x/evm/types"
	"github.com/FiboChain/fbc/x/order"
	tokencmd "github.com/FiboChain/fbc/x/token/client/cli"
	"github.com/spf13/cobra"
)

var (
	cdc = codec.MakeCodec(app.ModuleBasics)
)

func main() {
	// Configure cobra to sort commands
	cobra.EnableCommandSorting = false

	tmamino.RegisterKeyType(ethsecp256k1.PubKey{}, ethsecp256k1.PubKeyName)
	tmamino.RegisterKeyType(ethsecp256k1.PrivKey{}, ethsecp256k1.PrivKeyName)
	multisig.RegisterKeyType(ethsecp256k1.PubKey{}, ethsecp256k1.PubKeyName)

	keys.CryptoCdc = cdc
	clientkeys.KeysCdc = cdc

	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	fbchain.SetBech32Prefixes(config)
	fbchain.SetBip44CoinType(config)
	config.Seal()

	rootCmd := &cobra.Command{
		Use:   "fbchaincli",
		Short: "Command line interface for interacting with fbchaind",
	}

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(flags.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		utils.SetParseAppTx(parseMsgEthereumTx)
		return client.InitConfig(rootCmd)
	}

	// Construct Root Command
	rootCmd.AddCommand(
		clientrpc.StatusCommand(),
		sdkclient.ConfigCmd(app.DefaultCLIHome),
		queryCmd(cdc),
		txCmd(cdc),
		flags.LineBreak,
		client.KeyCommands(),
		client.AddrCommands(),
		flags.LineBreak,
		version.Cmd,
		flags.NewCompletionCmd(rootCmd, true),
	)

	// Add flags and prefix all env exposed with FIBOCHAIN
	executor := cli.PrepareMainCmd(rootCmd, "FIBOCHAIN", app.DefaultCLIHome)

	err := executor.Execute()
	if err != nil {
		panic(fmt.Errorf("failed executing CLI command: %w", err))
	}
}

func queryCmd(cdc *sdkcodec.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		authcmd.GetAccountCmd(cdc),
		flags.LineBreak,
		authcmd.QueryTxsByEventsCmd(cdc),
		authcmd.QueryTxCmd(cdc),
		flags.LineBreak,
	)

	// add modules' query commands
	app.ModuleBasics.AddQueryCommands(queryCmd, cdc)

	return queryCmd
}

func txCmd(cdc *sdkcodec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		tokencmd.SendTxCmd(cdc),
		flags.LineBreak,
		authcmd.GetSignCommand(cdc),
		authcmd.GetMultiSignCommand(cdc),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(cdc),
		authcmd.GetEncodeCommand(cdc),
		authcmd.GetDecodeCommand(cdc),
		flags.LineBreak,
	)

	// add modules' tx commands
	app.ModuleBasics.AddTxCommands(txCmd, cdc)

	// remove auth and bank commands as they're mounted under the root tx command
	var cmdsToRemove []*cobra.Command

	for _, cmd := range txCmd.Commands() {
		if cmd.Use == auth.ModuleName ||
			cmd.Use == order.ModuleName ||
			cmd.Use == dex.ModuleName ||
			cmd.Use == bank.ModuleName {
			cmdsToRemove = append(cmdsToRemove, cmd)
		}
	}

	txCmd.RemoveCommand(cmdsToRemove...)

	return txCmd
}

func parseMsgEthereumTx(cdc *sdkcodec.Codec, txBytes []byte) (sdk.Tx, error) {
	var tx evmtypes.MsgEthereumTx
	// try to decode through RLP first
	if err := authtypes.EthereumTxDecode(txBytes, &tx); err == nil {
		return &tx, nil
	}
	//try to decode through animo if it is not RLP-encoded
	if err := cdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}
