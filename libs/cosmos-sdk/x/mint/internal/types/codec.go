package types

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
)

const (
	ManageTreasuresProposalName = "fibochain/mint/ManageTreasuresProposal"
)

// ModuleCdc is a generic sealed codec to be used throughout this module
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	codec.RegisterCrypto(ModuleCdc)
	RegisterCodec(ModuleCdc)
	ModuleCdc.Seal()
}

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ManageTreasuresProposal{}, ManageTreasuresProposalName, nil)
}
