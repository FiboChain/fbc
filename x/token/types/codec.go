package types

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgTokenIssue{}, "fbchain/token/MsgIssue", nil)
	cdc.RegisterConcrete(MsgTokenBurn{}, "fbchain/token/MsgBurn", nil)
	cdc.RegisterConcrete(MsgTokenMint{}, "fbchain/token/MsgMint", nil)
	cdc.RegisterConcrete(MsgMultiSend{}, "fbchain/token/MsgMultiTransfer", nil)
	cdc.RegisterConcrete(MsgSend{}, "fbchain/token/MsgTransfer", nil)
	cdc.RegisterConcrete(MsgTransferOwnership{}, "fbchain/token/MsgTransferOwnership", nil)
	cdc.RegisterConcrete(MsgConfirmOwnership{}, "fbchain/token/MsgConfirmOwnership", nil)
	cdc.RegisterConcrete(MsgTokenModify{}, "fbchain/token/MsgModify", nil)

	// for test
	//cdc.RegisterConcrete(MsgTokenDestroy{}, "fbchain/token/MsgDestroy", nil)
}

// generic sealed codec to be used throughout this module
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
