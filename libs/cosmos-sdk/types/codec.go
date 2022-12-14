package types

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
)

// Register the sdk message type
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*Msg)(nil), nil)
	cdc.RegisterInterface((*Tx)(nil), nil)
	cdc.RegisterConcrete(BaseTx{}, "cosmos-sdk/BaseTx", nil)
}
