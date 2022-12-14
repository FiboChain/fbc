package types

import (
	"fmt"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	sdkerrors "github.com/FiboChain/fbc/libs/cosmos-sdk/types/errors"
	authtypes "github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/types"
	"github.com/FiboChain/fbc/libs/tendermint/global"
	"github.com/FiboChain/fbc/libs/tendermint/types"
)

const IGNORE_HEIGHT_CHECKING = -1

var errHeightLowerThanVenus = fmt.Errorf("lower than Venus")

// TxDecoder returns an sdk.TxDecoder that can decode both auth.StdTx and
// *MsgEthereumTx transactions.
func TxDecoder(cdc *codec.Codec) sdk.TxDecoder {
	return func(txBytes []byte, heights ...int64) (sdk.Tx, error) {
		if len(heights) > 1 {
			return nil, fmt.Errorf("to many height parameters")
		}
		var tx sdk.Tx
		var err error
		if len(txBytes) == 0 {
			return nil, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "tx bytes are empty")
		}

		var height int64
		if len(heights) == 1 {
			height = heights[0]
		} else {
			height = global.GetGlobalHeight()
		}

		for _, f := range []decodeFunc{
			evmDecoder,
			ubruDecoder,
			ubDecoder,
		} {
			if tx, err = f(cdc, txBytes, height); err == nil {
				switch realTx := tx.(type) {
				case authtypes.StdTx:
					realTx.Raw = txBytes
					realTx.Hash = types.Tx(txBytes).Hash(height)
					return realTx, nil
				case *MsgEthereumTx:
					realTx.Raw = txBytes
					realTx.Hash = types.Tx(txBytes).Hash(height)
					return realTx, nil
				}
			}
		}

		return nil, sdkerrors.Wrap(sdkerrors.ErrTxDecode, err.Error())
	}
}

type decodeFunc func(*codec.Codec, []byte, int64) (sdk.Tx, error)

// 1. Try to decode as MsgEthereumTx by RLP
func evmDecoder(_ *codec.Codec, txBytes []byte, height int64) (tx sdk.Tx, err error) {

	// bypass height checking in case of a negative number
	if height >= 0 && !types.HigherThanVenus(height) {
		err = errHeightLowerThanVenus
		return
	}

	var ethTx MsgEthereumTx
	if err = authtypes.EthereumTxDecode(txBytes, &ethTx); err == nil {
		tx = &ethTx
	}
	return
}

// 2. try customized unmarshalling implemented by UnmarshalFromAmino. higher performance!
func ubruDecoder(cdc *codec.Codec, txBytes []byte, height int64) (tx sdk.Tx, err error) {
	var v interface{}
	if v, err = cdc.UnmarshalBinaryLengthPrefixedWithRegisteredUbmarshaller(txBytes, &tx); err != nil {
		return nil, err
	}
	return sanityCheck(v.(sdk.Tx), height)
}

// TODO: switch to UnmarshalBinaryBare on SDK v0.40.0
// 3. the original amino way, decode by reflection.
func ubDecoder(cdc *codec.Codec, txBytes []byte, height int64) (tx sdk.Tx, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx)
	if err != nil {
		return nil, err
	}
	return sanityCheck(tx, height)
}

func sanityCheck(tx sdk.Tx, height int64) (sdk.Tx, error) {
	if tx.GetType() == sdk.EvmTxType && types.HigherThanVenus(height) {
		return nil, fmt.Errorf("amino decode is not allowed for MsgEthereumTx")
	}
	return tx, nil
}
