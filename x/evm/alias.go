package evm

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/x/evm/keeper"
	"github.com/FiboChain/fbc/x/evm/types"
)

// nolint
const (
	ModuleName        = types.ModuleName
	StoreKey          = types.StoreKey
	RouterKey         = types.RouterKey
	DefaultParamspace = types.DefaultParamspace
)

// nolint
var (
	NewKeeper            = keeper.NewKeeper
	TxDecoder            = types.TxDecoder
	NewSimulateKeeper    = keeper.NewSimulateKeeper
	NewLogProcessEvmHook = keeper.NewLogProcessEvmHook
	NewMultiEvmHooks     = keeper.NewMultiEvmHooks
)

//nolint
type (
	Keeper        = keeper.Keeper
	GenesisState  = types.GenesisState
	EvmLogHandler = types.EvmLogHandler
)

func WithMoreDeocder(cdc *codec.Codec, cc sdk.TxDecoder) sdk.TxDecoder {
	return func(txBytes []byte, height ...int64) (sdk.Tx, error) {
		ret, err := cc(txBytes, height...)
		if nil == err {
			return ret, nil
		}
		return ret, nil
	}
}
