package types

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreatePool{}, "fbchain/farm/MsgCreatePool", nil)
	cdc.RegisterConcrete(MsgDestroyPool{}, "fbchain/farm/MsgDestroyPool", nil)
	cdc.RegisterConcrete(MsgLock{}, "fbchain/farm/MsgLock", nil)
	cdc.RegisterConcrete(MsgUnlock{}, "fbchain/farm/MsgUnlock", nil)
	cdc.RegisterConcrete(MsgClaim{}, "fbchain/farm/MsgClaim", nil)
	cdc.RegisterConcrete(MsgProvide{}, "fbchain/farm/MsgProvide", nil)
	cdc.RegisterConcrete(ManageWhiteListProposal{}, "fbchain/farm/ManageWhiteListProposal", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
