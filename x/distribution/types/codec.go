package types

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgWithdrawValidatorCommission{}, "fbchain/distribution/MsgWithdrawReward", nil)
	cdc.RegisterConcrete(MsgSetWithdrawAddress{}, "fbchain/distribution/MsgModifyWithdrawAddress", nil)
	cdc.RegisterConcrete(CommunityPoolSpendProposal{}, "fbchain/distribution/CommunityPoolSpendProposal", nil)
}

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
