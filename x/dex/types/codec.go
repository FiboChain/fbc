package types

import "github.com/FiboChain/fbc/libs/cosmos-sdk/codec"

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgList{}, "fibochain/dex/MsgList", nil)
	cdc.RegisterConcrete(MsgDeposit{}, "fibochain/dex/MsgDeposit", nil)
	cdc.RegisterConcrete(MsgWithdraw{}, "fibochain/dex/MsgWithdraw", nil)
	cdc.RegisterConcrete(MsgTransferOwnership{}, "fibochain/dex/MsgTransferTradingPairOwnership", nil)
	cdc.RegisterConcrete(MsgConfirmOwnership{}, "fibochain/dex/MsgConfirmOwnership", nil)
	cdc.RegisterConcrete(DelistProposal{}, "fibochain/dex/DelistProposal", nil)
	cdc.RegisterConcrete(MsgCreateOperator{}, "fibochain/dex/CreateOperator", nil)
	cdc.RegisterConcrete(MsgUpdateOperator{}, "fibochain/dex/UpdateOperator", nil)
}

// ModuleCdc represents generic sealed codec to be used throughout this module
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
