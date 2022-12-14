package client

import (
	"encoding/hex"
	"fmt"
	"strings"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

const (
	fiboPrefix = "fbchain"
	fbPrefix   = "fb"
	rawPrefix  = "0x"
)

type accAddrToPrefixFunc func(sdk.AccAddress, string) string

// AddrCommands registers a sub-tree of commands to interact with oec address
func AddrCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addr",
		Short: "opreate all kind of address in the FBC network",
		Long: ` Address is a identification for join in the FBC network.

	The address in FBC network begins with "fbchain","fb" or "0x"`,
	}
	cmd.AddCommand(convertCommand())
	return cmd

}

func convertCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "convert [sourceAddr]",
		Short: "convert source address to all kind of address in the FBC network",
		Long: `sourceAddr must be begin with "fbchain","fb" or "0x".
	
	When input one of these address, we will convert to the other kinds.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addrList := make(map[string]string)
			targetPrefix := []string{fiboPrefix, fbPrefix, rawPrefix}
			srcAddr := args[0]

			// register func to encode account address to prefix address.
			toPrefixFunc := map[string]accAddrToPrefixFunc{
				fiboPrefix: bech32FromAccAddr,
				fbPrefix:   bech32FromAccAddr,
				rawPrefix:  hexFromAccAddr,
			}

			// prefix is "fbchain","fb" or "0x"
			// convert srcAddr to accAddr
			var accAddr sdk.AccAddress
			var err error
			switch {
			case strings.HasPrefix(srcAddr, fiboPrefix):
				//source address parse to account address
				addrList[fiboPrefix] = srcAddr
				accAddr, err = bech32ToAccAddr(fiboPrefix, srcAddr)

			case strings.HasPrefix(srcAddr, fbPrefix):
				//source address parse to account address
				addrList[fbPrefix] = srcAddr
				accAddr, err = bech32ToAccAddr(fbPrefix, srcAddr)

			case strings.HasPrefix(srcAddr, rawPrefix):
				addrList[rawPrefix] = srcAddr
				accAddr, err = hexToAccAddr(rawPrefix, srcAddr)

			default:
				return fmt.Errorf("unsupported prefix to convert")
			}

			// check account address
			if err != nil {
				fmt.Printf("Parse bech32 address error: %s", err)
				return err
			}

			// fill other kinds of prefix address out
			for _, pfx := range targetPrefix {
				if _, ok := addrList[pfx]; !ok {
					addrList[pfx] = toPrefixFunc[pfx](accAddr, pfx)
				}
			}

			//show all kinds of prefix address out
			for _, pfx := range targetPrefix {
				addrType := "Bech32"
				if pfx == "0x" {
					addrType = "Hex"
				}
				fmt.Printf("%s format with prefix <%s>: %5s\n", addrType, pfx, addrList[pfx])
			}

			return nil
		},
	}
}

// bech32ToAccAddr convert a hex string which begins with 'prefix' to an account address
func bech32ToAccAddr(prefix string, srcAddr string) (sdk.AccAddress, error) {
	return sdk.AccAddressFromBech32ByPrefix(srcAddr, prefix)
}

// bech32FromAccAddr create a hex string which begins with 'prefix' to from account address
func bech32FromAccAddr(accAddr sdk.AccAddress, prefix string) string {
	return accAddr.Bech32String(prefix)
}

// hexToAccAddr convert a hex string to an account address
func hexToAccAddr(prefix string, srcAddr string) (sdk.AccAddress, error) {
	srcAddr = strings.TrimPrefix(srcAddr, prefix)
	return sdk.AccAddressFromHex(srcAddr)
}

// hexFromAccAddr create a hex string from an account address
func hexFromAccAddr(accAddr sdk.AccAddress, prefix string) string {
	return prefix + hex.EncodeToString(accAddr.Bytes())
}