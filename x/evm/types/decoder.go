package types

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	sdkerrors "github.com/FiboChain/fbc/libs/cosmos-sdk/types/errors"
	typestx "github.com/FiboChain/fbc/libs/cosmos-sdk/types/tx"
	ibctxdecoder "github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/ibc-tx"
	authtypes "github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/types"
	"github.com/FiboChain/fbc/libs/tendermint/global"
	"github.com/FiboChain/fbc/libs/tendermint/types"
)

const IGNORE_HEIGHT_CHECKING = -1

// evmDecoder:  MsgEthereumTx decoder by Ethereum RLP
// ubruDecoder: customized unmarshalling implemented by UnmarshalFromAmino. higher performance!
// ubDecoder:   The original amino decoder, decoding by reflection
// ibcDecoder:  Protobuf decoder

// When and which decoder decoding what kind of tx:
// | ------------| --------------------|---------------|-------------|-----------------|----------------|
// |             | Before ubruDecoder  | Before Venus  | After Venus | Before VenusOne | After VenusOne |
// |             | carried out         |               |             |                 |                |
// | ------------|---------------------|---------------|-------------|-----------------|----------------|
// | evmDecoder  |                     |               |    evmtx    |   evmtx         |   evmtx        |
// | ubruDecoder |                     | stdtx & evmtx |    stdtx    |   stdtx         |   stdtx        |
// | ubDecoder   | stdtx,evmtx,otherTx | otherTx       |    otherTx  |   otherTx       |   otherTx      |
// | ibcDecoder  |                     |               |             |                 |   ibcTx        |
// | ------------| --------------------|---------------|-------------|-----------------|----------------|

func TxDecoder(cdc codec.CdcAbstraction) sdk.TxDecoder {

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

		for index, f := range []decodeFunc{
			evmDecoder,
			ubruDecoder,
			ubDecoder,
			ibcDecoder,
		} {
			if tx, err = f(cdc, txBytes, height); err == nil {
				tx.SetRaw(txBytes)
				tx.SetTxHash(types.Tx(txBytes).Hash(height))
				// index=0 means it is a evmtx(evmDecoder) ,we wont verify again
				// height > IGNORE_HEIGHT_CHECKING means it is a query request
				if index > 0 && height > IGNORE_HEIGHT_CHECKING {
					if sensitive, ok := tx.(sdk.HeightSensitive); ok {
						if err := sensitive.ValidWithHeight(height); err != nil {
							return nil, sdkerrors.Wrap(sdkerrors.ErrTxDecode, err.Error())
						}
					}
				}

				return tx, nil
			}
		}

		return nil, sdkerrors.Wrap(sdkerrors.ErrTxDecode, err.Error())
	}
}

// Unmarshaler is a generic type for Unmarshal functions
type Unmarshaler func(bytes []byte, ptr interface{}) error

func ibcDecoder(cdcWrapper codec.CdcAbstraction, bytes []byte, height int64) (tx sdk.Tx, err error) {
	if height >= 0 && !types.HigherThanVenus1(height) {
		err = fmt.Errorf("IbcTxDecoder decode tx err,lower than Venus1 height")
		return
	}
	simReq := &typestx.SimulateRequest{}
	txBytes := bytes

	err = simReq.Unmarshal(bytes)
	if err == nil && simReq.Tx != nil {
		txBytes, err = proto.Marshal(simReq.Tx)
		if err != nil {
			return nil, fmt.Errorf("relayTx invalid tx Marshal err %v", err)
		}
	}

	if txBytes == nil {
		return nil, errors.New("relayTx empty txBytes is not allowed")
	}

	cdc, ok := cdcWrapper.(*codec.CodecProxy)
	if !ok {
		return nil, errors.New("Invalid cdc abstraction!")
	}
	marshaler := cdc.GetProtocMarshal()
	decode := ibctxdecoder.IbcTxDecoder(marshaler)
	tx, err = decode(txBytes)
	if err != nil {
		return nil, fmt.Errorf("IbcTxDecoder decode tx err %v", err)
	}

	return
}

type decodeFunc func(codec.CdcAbstraction, []byte, int64) (sdk.Tx, error)

// 1. Try to decode as MsgEthereumTx by RLP
func evmDecoder(_ codec.CdcAbstraction, txBytes []byte, height int64) (tx sdk.Tx, err error) {

	// bypass height checking in case of a negative number
	if height >= 0 && !types.HigherThanVenus(height) {
		err = fmt.Errorf("lower than Venus")
		return
	}

	var ethTx MsgEthereumTx
	if err = authtypes.EthereumTxDecode(txBytes, &ethTx); err == nil {
		tx = &ethTx
	}
	return
}

// 2. try customized unmarshalling implemented by UnmarshalFromAmino. higher performance!
func ubruDecoder(cdc codec.CdcAbstraction, txBytes []byte, height int64) (tx sdk.Tx, err error) {
	var v interface{}
	if v, err = cdc.UnmarshalBinaryLengthPrefixedWithRegisteredUbmarshaller(txBytes, &tx); err != nil {
		return nil, err
	}
	return sanityCheck(v.(sdk.Tx), height)
}

// TODO: switch to UnmarshalBinaryBare on SDK v0.40.0
// 3. the original amino way, decode by reflection.
func ubDecoder(cdc codec.CdcAbstraction, txBytes []byte, height int64) (tx sdk.Tx, err error) {
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
