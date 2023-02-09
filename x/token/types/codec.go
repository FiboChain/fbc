package types

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgTokenIssue{}, "fibochain/token/MsgIssue", nil)
	cdc.RegisterConcrete(MsgTokenBurn{}, "fibochain/token/MsgBurn", nil)
	cdc.RegisterConcrete(MsgTokenMint{}, "fibochain/token/MsgMint", nil)
	cdc.RegisterConcrete(MsgMultiSend{}, "fibochain/token/MsgMultiTransfer", nil)
	cdc.RegisterConcrete(MsgSend{}, "fibochain/token/MsgTransfer", nil)
	cdc.RegisterConcrete(MsgTransferOwnership{}, "fibochain/token/MsgTransferOwnership", nil)
	cdc.RegisterConcrete(MsgConfirmOwnership{}, "fibochain/token/MsgConfirmOwnership", nil)
	cdc.RegisterConcrete(MsgTokenModify{}, "fibochain/token/MsgModify", nil)

	// for test
	//cdc.RegisterConcrete(MsgTokenDestroy{}, "fibochain/token/MsgDestroy", nil)
}

// generic sealed codec to be used throughout this module
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
