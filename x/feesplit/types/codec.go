package types

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
)

// ModuleCdc defines the feesplit module's codec
var ModuleCdc = codec.New()

const (
	// Amino names
	registerFeeSplitName = "fibochain/MsgRegisterFeeSplit"
	updateFeeSplitName   = "fibochain/MsgUpdateFeeSplit"
	cancelFeeSplitName   = "fibochain/MsgCancelFeeSplit"
	sharesProposalName   = "fibochain/feesplit/SharesProposal"
)

// NOTE: This is required for the GetSignBytes function
func init() {
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}

// RegisterCodec registers all the necessary types and interfaces for the
// feesplit module
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgRegisterFeeSplit{}, registerFeeSplitName, nil)
	cdc.RegisterConcrete(MsgUpdateFeeSplit{}, updateFeeSplitName, nil)
	cdc.RegisterConcrete(MsgCancelFeeSplit{}, cancelFeeSplitName, nil)
	cdc.RegisterConcrete(FeeSplitSharesProposal{}, sharesProposalName, nil)
}
