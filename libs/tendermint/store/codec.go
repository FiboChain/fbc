package store

import (
	amino "github.com/tendermint/go-amino"

	"github.com/FiboChain/fbc/libs/tendermint/types"
)

var cdc = amino.NewCodec()

func init() {
	types.RegisterBlockAmino(cdc)
}
