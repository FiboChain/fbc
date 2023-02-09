package types

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreatePool{}, "fibochain/farm/MsgCreatePool", nil)
	cdc.RegisterConcrete(MsgDestroyPool{}, "fibochain/farm/MsgDestroyPool", nil)
	cdc.RegisterConcrete(MsgLock{}, "fibochain/farm/MsgLock", nil)
	cdc.RegisterConcrete(MsgUnlock{}, "fibochain/farm/MsgUnlock", nil)
	cdc.RegisterConcrete(MsgClaim{}, "fibochain/farm/MsgClaim", nil)
	cdc.RegisterConcrete(MsgProvide{}, "fibochain/farm/MsgProvide", nil)
	cdc.RegisterConcrete(ManageWhiteListProposal{}, "fibochain/farm/ManageWhiteListProposal", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
