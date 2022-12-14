package ante_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth"

	"github.com/FiboChain/fbc/app"
	ante "github.com/FiboChain/fbc/app/ante"
	appconfig "github.com/FiboChain/fbc/app/config"
	"github.com/FiboChain/fbc/app/crypto/ethsecp256k1"
	fbchain "github.com/FiboChain/fbc/app/types"
	evmtypes "github.com/FiboChain/fbc/x/evm/types"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	abci "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	tmcrypto "github.com/FiboChain/fbc/libs/tendermint/crypto"
)

type AnteTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.FBchainApp
	anteHandler sdk.AnteHandler
}

func (suite *AnteTestSuite) SetupTest() {
	checkTx := false
	chainId := "fbchain-3"

	suite.app = app.Setup(checkTx)
	suite.app.Codec().RegisterConcrete(&sdk.TestMsg{}, "test/TestMsg", nil)

	suite.ctx = suite.app.BaseApp.NewContext(checkTx, abci.Header{Height: 1, ChainID: chainId, Time: time.Now().UTC()})
	suite.app.EvmKeeper.SetParams(suite.ctx, evmtypes.DefaultParams())

	suite.anteHandler = ante.NewAnteHandler(suite.app.AccountKeeper, suite.app.EvmKeeper, suite.app.SupplyKeeper, nil)

	err := fbchain.SetChainId(chainId)
	suite.Nil(err)

	appconfig.RegisterDynamicConfig(suite.app.Logger())
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

func newTestMsg(addrs ...sdk.AccAddress) *sdk.TestMsg {
	return sdk.NewTestMsg(addrs...)
}

func newTestCoins() sdk.Coins {
	return sdk.NewCoins(fbchain.NewPhotonCoinInt64(500000000))
}

func newTestStdFee() auth.StdFee {
	return auth.NewStdFee(220000, sdk.NewCoins(fbchain.NewPhotonCoinInt64(150)))
}

// GenerateAddress generates an Ethereum address.
func newTestAddrKey() (sdk.AccAddress, tmcrypto.PrivKey) {
	privkey, _ := ethsecp256k1.GenerateKey()
	addr := ethcrypto.PubkeyToAddress(privkey.ToECDSA().PublicKey)

	return sdk.AccAddress(addr.Bytes()), privkey
}

func newTestSDKTx(
	ctx sdk.Context, msgs []sdk.Msg, privs []tmcrypto.PrivKey,
	accNums []uint64, seqs []uint64, fee auth.StdFee,
) sdk.Tx {

	sigs := make([]auth.StdSignature, len(privs))
	for i, priv := range privs {
		signBytes := auth.StdSignBytes(ctx.ChainID(), accNums[i], seqs[i], fee, msgs, "")

		sig, err := priv.Sign(signBytes)
		if err != nil {
			panic(err)
		}

		sigs[i] = auth.StdSignature{
			PubKey:    priv.PubKey(),
			Signature: sig,
		}
	}

	return auth.NewStdTx(msgs, fee, sigs, "")
}

func newTestEthTx(ctx sdk.Context, msg *evmtypes.MsgEthereumTx, priv tmcrypto.PrivKey) (sdk.Tx, error) {
	chainIDEpoch, err := fbchain.ParseChainID(ctx.ChainID())
	if err != nil {
		return nil, err
	}

	privkey, ok := priv.(ethsecp256k1.PrivKey)
	if !ok {
		return nil, fmt.Errorf("invalid private key type: %T", priv)
	}

	if err := msg.Sign(chainIDEpoch, privkey.ToECDSA()); err != nil {
		return nil, err
	}

	return msg, nil
}
