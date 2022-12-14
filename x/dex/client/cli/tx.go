package cli

import (
	"bufio"
	"fmt"
	"strings"

	client "github.com/FiboChain/fbc/libs/cosmos-sdk/client/flags"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/version"
	"github.com/FiboChain/fbc/x/gov"

	"github.com/pkg/errors"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/client/context"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth"
	authTypes "github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/client/utils"
	"github.com/FiboChain/fbc/x/common"
	dexUtils "github.com/FiboChain/fbc/x/dex/client/utils"
	"github.com/FiboChain/fbc/x/dex/types"
	"github.com/spf13/cobra"
)

// Dex tags
const (
	FlagBaseAsset          = "base-asset"
	FlagQuoteAsset         = "quote-asset"
	FlagInitPrice          = "init-price"
	FlagProduct            = "product"
	FlagFrom               = "from"
	FlagTo                 = "to"
	FlagWebsite            = "website"
	FlagHandlingFeeAddress = "handling-fee-address"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "dex",
		Short: "Decentralized exchange management subcommands",
	}

	txCmd.AddCommand(client.PostCommands(
		getCmdList(cdc),
		getCmdDeposit(cdc),
		getCmdWithdraw(cdc),
		getCmdTransferOwnership(cdc),
		getCmdConfirmOwnership(cdc),
		getCmdRegisterOperator(cdc),
		getCmdEditOperator(cdc),
	)...)

	return txCmd
}

func getCmdList(cdc *codec.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list a trading pair",
		Args:  cobra.ExactArgs(0),
		Long: strings.TrimSpace(`List a trading pair:

$ fbchaincli tx dex list --base-asset mytoken --quote-asset fibo --from mykey
`),
		RunE: func(cmd *cobra.Command, _ []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			if err := auth.NewAccountRetriever(cliCtx).EnsureExists(cliCtx.FromAddress); err != nil {
				return err
			}

			flags := cmd.Flags()
			baseAsset, err := flags.GetString(FlagBaseAsset)
			if err != nil {
				return err
			}
			if len(baseAsset) == 0 {
				return errors.New("failed. empty base asset")
			}
			quoteAsset, err := flags.GetString(FlagQuoteAsset)
			if err != nil {
				return err
			}
			strInitPrice, err := flags.GetString(FlagInitPrice)
			if err != nil {
				return err
			}
			initPrice := sdk.MustNewDecFromStr(strInitPrice)
			owner := cliCtx.GetFromAddress()
			listMsg := types.NewMsgList(owner, baseAsset, quoteAsset, initPrice)
			return utils.CompleteAndBroadcastTxCLI(txBldr, cliCtx, []sdk.Msg{listMsg})
		},
	}

	cmd.Flags().StringP(FlagBaseAsset, "", "", FlagBaseAsset+" should be issued before listed to opendex")
	cmd.Flags().StringP(FlagQuoteAsset, "", common.NativeToken, FlagQuoteAsset+" should be issued before listed to opendex")
	cmd.Flags().StringP(FlagInitPrice, "", "0.01", FlagInitPrice+" should be valid price")

	return cmd
}

// getCmdDeposit implements depositing tokens for a product.
func getCmdDeposit(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "deposit [product] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "deposit an amount of token on a product",
		Long: strings.TrimSpace(`Deposit an amount of token on a product:

$ fbchaincli tx dex deposit mytoken_fibo 1000fibo --from mykey

The 'product' is a trading pair in full name of the tokens: ${base-asset-symbol}_${quote-asset-symbol}, for example 'mytoken_fibo'.
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			product := args[0]

			// Get depositor address
			from := cliCtx.GetFromAddress()

			// Get amount of coins
			amount, err := sdk.ParseDecCoin(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgDeposit(product, amount, from)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// getCmdWithdraw implements withdrawing tokens from a product.
func getCmdWithdraw(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw [product] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "withdraw an amount of token from a product",
		Long: strings.TrimSpace(`Withdraw an amount of token from a product:

$ fbchaincli tx dex withdraw mytoken_fibo 1000fibo --from mykey

The 'product' is a trading pair in full name of the tokens: ${base-asset-symbol}_${quote-asset-symbol}, for example 'mytoken_fibo'.
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			product := args[0]

			// Get depositor address
			from := cliCtx.GetFromAddress()

			// Get amount of coins
			amount, err := sdk.ParseDecCoin(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgWithdraw(product, amount, from)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// getCmdTransferOwnership is the CLI command for transfer ownership of product
func getCmdTransferOwnership(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-ownership",
		Short: "change the owner of the product",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			if err := authTypes.NewAccountRetriever(cliCtx).EnsureExists(cliCtx.FromAddress); err != nil {
				return err
			}
			flags := cmd.Flags()

			product, err := flags.GetString(FlagProduct)
			if err != nil || product == "" {
				return fmt.Errorf("invalid product:%s", product)
			}

			to, err := flags.GetString(FlagTo)
			if err != nil {
				return fmt.Errorf("invalid to:%s", to)
			}

			toAddr, err := sdk.AccAddressFromBech32(to)
			if err != nil {
				return fmt.Errorf("invalid to:%s", to)
			}

			from := cliCtx.GetFromAddress()
			msg := types.NewMsgTransferOwnership(from, toAddr, product)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().StringP(FlagProduct, "p", "", "product to be transferred")
	cmd.Flags().String(FlagTo, "", "the user to be transferred")
	return cmd
}

func getCmdConfirmOwnership(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm-ownership",
		Short: "confirm the transfer-ownership of the product",
		RunE: func(cmd *cobra.Command, args []string) error {

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			if err := authTypes.NewAccountRetriever(cliCtx).EnsureExists(cliCtx.FromAddress); err != nil {
				return err
			}
			flags := cmd.Flags()

			product, err := flags.GetString(FlagProduct)
			if err != nil {
				return errors.New("invalid product")
			}
			_, err = flags.GetString(FlagFrom)
			if err != nil {
				return errors.New("invalid from")
			}

			from := cliCtx.GetFromAddress()

			msg := types.NewMsgConfirmOwnership(from, product)
			return utils.CompleteAndBroadcastTxCLI(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}
	cmd.Flags().StringP(FlagProduct, "p", "", "product to be transferred")
	return cmd
}

// GetCmdSubmitDelistProposal implememts a command handler for submitting a dex delist proposal transaction
func GetCmdSubmitDelistProposal(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "delist-proposal [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a dex delist proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a dex delist proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal delist-proposal <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
 "title": "delist xxx/%s",
 "description": "delist asset from dex",
 "base_asset": "xxx",
 "quote_asset": "%s",
 "deposit": [
   {
     "denom": "%s",
     "amount": "100"
   }
 ]
}
`, version.ClientName, sdk.DefaultBondDenom, sdk.DefaultBondDenom, sdk.DefaultBondDenom,
			)),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			proposal, err := dexUtils.ParseDelistProposalJSON(cdc, args[0])
			if err != nil {
				return err
			}

			from := cliCtx.GetFromAddress()
			content := types.NewDelistProposal(proposal.Title, proposal.Description, from, proposal.BaseAsset, proposal.QuoteAsset)
			msg := gov.NewMsgSubmitProposal(content, proposal.Deposit, from)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

}

func getCmdRegisterOperator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-operator",
		Short: "register a dex operator",
		Args:  cobra.ExactArgs(0),
		Long: strings.TrimSpace(`Register a dex operator:

$ fbchaincli tx dex register-operator --website http://xxx/operator.json --handling-fee-address addr --from mykey
`),
		RunE: func(cmd *cobra.Command, _ []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			if err := auth.NewAccountRetriever(cliCtx).EnsureExists(cliCtx.FromAddress); err != nil {
				return err
			}

			flags := cmd.Flags()
			website, err := flags.GetString(FlagWebsite)
			if err != nil {
				return err
			}
			feeAddrStr, err := flags.GetString(FlagHandlingFeeAddress)
			if err != nil {
				return err
			}
			feeAddr, err := sdk.AccAddressFromBech32(feeAddrStr)
			if err != nil {
				return sdk.ErrInvalidAddress(fmt.Sprintf("invalid address：%s", feeAddrStr))
			}
			owner := cliCtx.GetFromAddress()
			operatorMsg := types.NewMsgCreateOperator(website, owner, feeAddr)
			return utils.CompleteAndBroadcastTxCLI(txBldr, cliCtx, []sdk.Msg{operatorMsg})
		},
	}

	cmd.Flags().String(FlagWebsite, "", `A valid http link to describe DEXOperator which ends with "operator.json" defined in OIP-{xxx}，and its length should be less than 1024`)
	cmd.Flags().String(FlagHandlingFeeAddress, "", "An address to receive fees of tokenpair's matched order")

	return cmd
}

func getCmdEditOperator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-operator",
		Short: "edit a dex operator",
		Args:  cobra.ExactArgs(0),
		Long: strings.TrimSpace(`Edit a dex operator:

$ fbchaincli tx dex edit-operator --website http://xxx/operator.json --handling-fee-address addr --from mykey
`),
		RunE: func(cmd *cobra.Command, _ []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			if err := auth.NewAccountRetriever(cliCtx).EnsureExists(cliCtx.FromAddress); err != nil {
				return err
			}

			flags := cmd.Flags()
			website, err := flags.GetString(FlagWebsite)
			if err != nil {
				return err
			}
			feeAddrStr, err := flags.GetString(FlagHandlingFeeAddress)
			if err != nil {
				return err
			}
			feeAddr, err := sdk.AccAddressFromBech32(feeAddrStr)
			if err != nil {
				return sdk.ErrInvalidAddress(fmt.Sprintf("invalid address：%s", feeAddrStr))
			}
			owner := cliCtx.GetFromAddress()
			operatorMsg := types.NewMsgUpdateOperator(website, owner, feeAddr)
			return utils.CompleteAndBroadcastTxCLI(txBldr, cliCtx, []sdk.Msg{operatorMsg})
		},
	}

	cmd.Flags().String(FlagWebsite, "", `A valid http link to describe DEXOperator which ends with "operator.json" defined in OIP-{xxx}，and its length should be less than 1024`)
	cmd.Flags().String(FlagHandlingFeeAddress, "", "An address to receive fees of tokenpair's matched order")

	return cmd
}
