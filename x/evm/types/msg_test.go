package types

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"

	"github.com/stretchr/testify/require"

	"github.com/FiboChain/fbc/app/crypto/ethsecp256k1"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/FiboChain/fbc/libs/tendermint/crypto/secp256k1"
)

func newSdkAddress() sdk.AccAddress {
	tmpKey := secp256k1.GenPrivKey().PubKey()
	return sdk.AccAddress(tmpKey.Address().Bytes())
}

func TestMsgEthereumTx(t *testing.T) {
	addr := GenerateEthAddress()

	msg := NewMsgEthereumTx(0, &addr, nil, 100000, nil, []byte("test"))
	require.NotNil(t, msg)
	require.Equal(t, *msg.Data.Recipient, addr)
	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), TypeMsgEthereumTx)
	require.NotNil(t, msg.To())
	require.Equal(t, msg.GetMsgs(), []sdk.Msg{msg})
	require.Panics(t, func() { msg.GetSigners() })
	require.Panics(t, func() { msg.GetSignBytes() })

	msg = NewMsgEthereumTxContract(0, nil, 100000, nil, []byte("test"))
	require.NotNil(t, msg)
	require.Nil(t, msg.Data.Recipient)
	require.Nil(t, msg.To())
}

func TestMsgEthereumTxValidation(t *testing.T) {
	testCases := []struct {
		msg        string
		amount     *big.Int
		gasPrice   *big.Int
		expectPass bool
	}{
		{msg: "pass", amount: big.NewInt(100), gasPrice: big.NewInt(100000), expectPass: true},
		{msg: "pass amount is zero", amount: big.NewInt(0), gasPrice: big.NewInt(100000), expectPass: true},
		{msg: "invalid amount", amount: big.NewInt(-1), gasPrice: big.NewInt(100000), expectPass: false},
		{msg: "invalid gas price", amount: big.NewInt(100), gasPrice: big.NewInt(-1), expectPass: false},
		{msg: "invalid gas price", amount: big.NewInt(100), gasPrice: big.NewInt(0), expectPass: false},
	}

	for i, tc := range testCases {
		msg := NewMsgEthereumTx(0, nil, tc.amount, 0, tc.gasPrice, nil)

		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "valid test %d failed: %s", i, tc.msg)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "invalid test %d passed: %s", i, tc.msg)
		}
	}
}

func TestMsgEthereumTxRLPSignBytes(t *testing.T) {
	addr := ethcmn.BytesToAddress([]byte("test_address"))
	chainID := big.NewInt(3)

	msg := NewMsgEthereumTx(0, &addr, nil, 100000, nil, []byte("test"))
	hash := msg.RLPSignBytes(chainID)
	require.Equal(t, "5BD30E35AD27449390B14C91E6BCFDCAADF8FE44EF33680E3BC200FC0DC083C7", fmt.Sprintf("%X", hash))
}

func TestMsgEthereumTxRLPEncode(t *testing.T) {
	addr := ethcmn.BytesToAddress([]byte("test_address"))
	msg := NewMsgEthereumTx(0, &addr, nil, 100000, nil, []byte("test"))

	raw, err := rlp.EncodeToBytes(&msg)
	require.NoError(t, err)
	require.Equal(t, ethcmn.FromHex("E48080830186A0940000000000000000746573745F61646472657373808474657374808080"), raw)
}

func TestMsgEthereumTxRLPDecode(t *testing.T) {
	var msg MsgEthereumTx

	raw := ethcmn.FromHex("E48080830186A0940000000000000000746573745F61646472657373808474657374808080")
	addr := ethcmn.BytesToAddress([]byte("test_address"))
	expectedMsg := NewMsgEthereumTx(0, &addr, nil, 100000, nil, []byte("test"))

	err := rlp.Decode(bytes.NewReader(raw), &msg)
	require.NoError(t, err)
	require.Equal(t, expectedMsg.Data, msg.Data)

	// value size exceeds available input length of stream
	mockStream := rlp.NewStream(bytes.NewReader(raw), 1)
	require.Error(t, msg.DecodeRLP(mockStream))
}

func TestMsgEthereumTxSig(t *testing.T) {
	chainID, zeroChainID := big.NewInt(3), big.NewInt(0)

	priv1, _ := ethsecp256k1.GenerateKey()
	priv2, _ := ethsecp256k1.GenerateKey()
	addr1 := ethcmn.BytesToAddress(priv1.PubKey().Address().Bytes())
	trimed := strings.TrimPrefix(addr1.Hex(), "0x")

	fmt.Printf("%s\n", trimed)
	addrSDKAddr1, err := sdk.AccAddressFromHex(trimed)
	require.NoError(t, err)
	addr2 := ethcmn.BytesToAddress(priv2.PubKey().Address().Bytes())

	// require valid signature passes validation
	msg := NewMsgEthereumTx(0, &addr1, nil, 100000, nil, []byte("test"))
	err = msg.Sign(chainID, priv1.ToECDSA())
	require.Nil(t, err)

	err = msg.VerifySig(chainID, 0)
	require.NoError(t, err)
	require.Equal(t, addr1, msg.EthereumAddress())
	require.NotEqual(t, addr2, msg.EthereumAddress())

	signers := msg.GetSigners()
	require.Equal(t, 1, len(signers))
	require.True(t, addrSDKAddr1.Equals(signers[0]))

	// zero chainID
	err = msg.Sign(zeroChainID, priv1.ToECDSA())
	require.Nil(t, err)
	err = msg.VerifySig(zeroChainID, 0)
	require.Nil(t, err)

	// require invalid chain ID fail validation
	msg = NewMsgEthereumTx(0, &addr1, nil, 100000, nil, []byte("test"))
	err = msg.Sign(chainID, priv1.ToECDSA())
	require.Nil(t, err)
}

func TestMsgEthereumTx_ChainID(t *testing.T) {
	chainID := big.NewInt(3)
	priv, _ := ethsecp256k1.GenerateKey()
	addr := ethcmn.BytesToAddress(priv.PubKey().Address().Bytes())
	msg := NewMsgEthereumTx(0, &addr, nil, 100000, nil, []byte("test"))
	err := msg.Sign(chainID, priv.ToECDSA())
	require.Nil(t, err)

	require.True(t, chainID.Cmp(msg.ChainID()) == 0)

	msg.Data.V = big.NewInt(27)
	require.NotNil(t, msg.ChainID())

	msg.Data.V = math.MaxBig256
	expectedChainID := new(big.Int).Div(new(big.Int).Sub(math.MaxBig256, big.NewInt(35)), big.NewInt(2))
	require.True(t, expectedChainID.Cmp(msg.ChainID()) == 0)
}

func TestMsgEthereumTxGetter(t *testing.T) {
	priv, _ := ethsecp256k1.GenerateKey()
	addr := ethcmn.BytesToAddress(priv.PubKey().Address().Bytes())
	amount, gasPrice, gasLimit := int64(1024), int64(2048), uint64(100000)
	expectedFee := gasPrice * int64(gasLimit)
	expectCost := expectedFee + amount
	msg := NewMsgEthereumTx(0, &addr, big.NewInt(amount), gasLimit, big.NewInt(gasPrice), []byte("test"))

	require.Equal(t, gasLimit, msg.GetGas())
	require.True(t, big.NewInt(expectedFee).Cmp(msg.Fee()) == 0)
	require.True(t, big.NewInt(expectCost).Cmp(msg.Cost()) == 0)

	expectedV, expectedR, expectedS := big.NewInt(1), big.NewInt(2), big.NewInt(3)
	msg.Data.V, msg.Data.R, msg.Data.S = expectedV, expectedR, expectedS
	v, r, s := msg.RawSignatureValues()
	require.True(t, expectedV.Cmp(v) == 0)
	require.True(t, expectedR.Cmp(r) == 0)
	require.True(t, expectedS.Cmp(s) == 0)
}

func TestMsgEthereumTx_Amino(t *testing.T) {
	priv, _ := ethsecp256k1.GenerateKey()
	addr := ethcmn.BytesToAddress(priv.PubKey().Address().Bytes())
	amount, gasPrice, gasLimit := int64(1024), int64(2048), uint64(100000)
	msg := NewMsgEthereumTx(0, &addr, big.NewInt(amount), gasLimit, big.NewInt(gasPrice), []byte("test"))
	err := msg.Sign(big.NewInt(3), priv.ToECDSA())
	require.NoError(t, err)
	hash := ethcmn.BigToHash(big.NewInt(2))

	testCases := []*MsgEthereumTx{
		msg,
		{
			Data: TxData{
				AccountNonce: 2,
				Price:        big.NewInt(3),
				GasLimit:     1,
				Recipient:    &addr,
				Amount:       big.NewInt(4),
				Payload:      []byte("test"),
				V:            big.NewInt(5),
				R:            big.NewInt(6),
				S:            big.NewInt(7),
				Hash:         &hash,
			},
		},
		{
			Data: TxData{
				Price:     big.NewInt(math.MinInt64),
				Recipient: &ethcmn.Address{},
				Amount:    big.NewInt(math.MinInt64),
				Payload:   []byte{},
				V:         big.NewInt(math.MinInt64),
				R:         big.NewInt(math.MinInt64),
				S:         big.NewInt(math.MinInt64),
				Hash:      &ethcmn.Hash{},
			},
		},
		{
			Data: TxData{
				AccountNonce: math.MaxUint64,
				Price:        big.NewInt(math.MaxInt64),
				GasLimit:     math.MaxUint64,
				Amount:       big.NewInt(math.MaxInt64),
				V:            big.NewInt(math.MaxInt64),
				R:            big.NewInt(math.MaxInt64),
				S:            big.NewInt(math.MaxInt64),
			},
		},
	}

	for _, msg := range testCases {
		raw, err := ModuleCdc.MarshalBinaryBare(msg)
		require.NoError(t, err)

		var msg2 MsgEthereumTx
		err = ModuleCdc.UnmarshalBinaryBare(raw, &msg2)
		require.NoError(t, err)

		var msg3 MsgEthereumTx
		v, err := ModuleCdc.UnmarshalBinaryBareWithRegisteredUnmarshaller(raw, &msg3)
		require.NoError(t, err)
		msg3 = *v.(*MsgEthereumTx)
		require.EqualValues(t, msg2, msg3)
	}
}

func BenchmarkMsgEthereumTxUnmarshal(b *testing.B) {
	cdc := ModuleCdc
	priv, _ := ethsecp256k1.GenerateKey()
	addr := ethcmn.BytesToAddress(priv.PubKey().Address().Bytes())
	amount, gasPrice, gasLimit := int64(1024), int64(2048), uint64(100000)
	msg := NewMsgEthereumTx(123456, &addr, big.NewInt(amount), gasLimit, big.NewInt(gasPrice), []byte("test"))
	_ = msg.Sign(big.NewInt(66), priv.ToECDSA())

	raw, _ := cdc.MarshalBinaryBare(msg)
	rlpRaw, err := rlp.EncodeToBytes(&msg)
	require.NoError(b, err)
	b.ResetTimer()

	b.Run("amino", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var msg2 MsgEthereumTx
			err := cdc.UnmarshalBinaryBare(raw, &msg2)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("unmarshaler", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var msg3 MsgEthereumTx
			v, err := cdc.UnmarshalBinaryBareWithRegisteredUnmarshaller(raw, &msg3)
			if err != nil {
				b.Fatal(err)
			}
			msg3 = v.(MsgEthereumTx)
		}
	})

	b.Run("rlp", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var msg MsgEthereumTx
			err = rlp.DecodeBytes(rlpRaw, &msg)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestMarshalAndUnmarshalLogs(t *testing.T) {
	var cdc = codec.New()

	logs := []*ethtypes.Log{
		{
			Address: ethcmn.BytesToAddress([]byte{0x11}),
			TxHash:  ethcmn.HexToHash("0x01"),
			// May need to find workaround since Topics is required to unmarshal from JSON
			Topics:  []ethcmn.Hash{},
			Removed: true,
		},
		{Address: ethcmn.BytesToAddress([]byte{0x01, 0x11}), Topics: []ethcmn.Hash{}},
	}

	raw, err := codec.MarshalJSONIndent(cdc, logs)
	require.NoError(t, err)

	var logs2 []*ethtypes.Log
	err = cdc.UnmarshalJSON(raw, &logs2)
	require.NoError(t, err)

	require.Len(t, logs2, 2)
	require.Equal(t, logs[0].Address, logs2[0].Address)
	require.Equal(t, logs[0].TxHash, logs2[0].TxHash)
	require.True(t, logs[0].Removed)

	emptyLogs := []*ethtypes.Log{}

	raw, err = codec.MarshalJSONIndent(cdc, emptyLogs)
	require.NoError(t, err)

	err = cdc.UnmarshalJSON(raw, &logs2)
	require.NoError(t, err)
}

func TestMsgString(t *testing.T) {
	expectedUint64, expectedSDKAddr := uint64(1024), newSdkAddress()
	expectedPayload, err := hexutil.Decode("0x1234567890abcdef")
	require.NoError(t, err)
	expectedOutput := fmt.Sprintf("nonce=1024 gasPrice=1 gasLimit=1024 recipient=%s amount=1 data=0x1234567890abcdef from=%s",
		expectedSDKAddr, expectedSDKAddr)

	expectedHexAddr := ethcmn.BytesToAddress([]byte{0x01})
	expectedBigInt := big.NewInt(1024)
	expectedOutput = fmt.Sprintf("nonce=1024 price=1024 gasLimit=1024 recipient=%s amount=1024 data=0x1234567890abcdef v=0 r=0 s=0", expectedHexAddr.Hex())
	msgEthereumTx := NewMsgEthereumTx(expectedUint64, &expectedHexAddr, expectedBigInt, expectedUint64, expectedBigInt, expectedPayload)
	require.True(t, strings.EqualFold(msgEthereumTx.String(), expectedOutput))
}
