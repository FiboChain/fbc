package app

import (
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/suite"

	appconfig "github.com/FiboChain/fbc/app/config"
	"github.com/FiboChain/fbc/app/crypto/ethsecp256k1"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	cosmossdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	authclient "github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/client/utils"
	abcitypes "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	"github.com/FiboChain/fbc/libs/tendermint/global"
	tendertypes "github.com/FiboChain/fbc/libs/tendermint/types"
	evmtypes "github.com/FiboChain/fbc/x/evm/types"
)

var (
	txCoin100000 = cosmossdk.NewInt64Coin(cosmossdk.DefaultBondDenom, 100000)
	nonce        = uint64(0)
)

type FakeBlockRecommendGPTestSuite struct {
	suite.Suite
	app   *FBChainApp
	codec *codec.Codec

	evmSenderPrivKey   ethsecp256k1.PrivKey
	evmContractAddress ethcommon.Address
}

func (suite *FakeBlockRecommendGPTestSuite) SetupTest() {
	suite.app = Setup(checkTx, WithChainId(cosmosChainId))
	suite.codec = suite.app.Codec()
	params := evmtypes.DefaultParams()
	params.EnableCreate = true
	params.EnableCall = true
	suite.app.EvmKeeper.SetParams(suite.Ctx(), params)

	suite.evmSenderPrivKey, _ = ethsecp256k1.GenerateKey()
	suite.evmContractAddress = ethcrypto.CreateAddress(ethcommon.HexToAddress(suite.evmSenderPrivKey.PubKey().Address().String()), 0)
	accountEvm := suite.app.AccountKeeper.NewAccountWithAddress(suite.Ctx(), suite.evmSenderPrivKey.PubKey().Address().Bytes())
	accountEvm.SetAccountNumber(accountNum)
	accountEvm.SetCoins(cosmossdk.NewCoins(txCoin100000))
	suite.app.AccountKeeper.SetAccount(suite.Ctx(), accountEvm)
}

func (suite *FakeBlockRecommendGPTestSuite) Ctx() cosmossdk.Context {
	return suite.app.BaseApp.GetDeliverStateCtx()
}

func (suite *FakeBlockRecommendGPTestSuite) beginFakeBlock(height int64) {
	tendertypes.UnittestOnlySetMilestoneVenusHeight(height - 1)
	global.SetGlobalHeight(height - 1)
	suite.app.BeginBlocker(suite.Ctx(), abcitypes.RequestBeginBlock{Header: abcitypes.Header{Height: height}})
}

func (suite *FakeBlockRecommendGPTestSuite) endFakeBlock(recommendGp string) {
	suite.app.EndBlocker(suite.Ctx(), abcitypes.RequestEndBlock{})
	ctx := suite.Ctx()
	suite.Require().True(recommendGp == GlobalGp.String(), "recommend gas price expect %s, but %s ", recommendGp, GlobalGp.String())
	//fmt.Println("GlobalGp: ", GlobalGp)
	suite.Require().False(ctx.BlockGasMeter().IsPastLimit())
	suite.Require().False(ctx.BlockGasMeter().IsOutOfGas())
}

func generateEvmTxs(totalTxNum int, baseGP int64, gpOffset *int64, decreaseGP bool, needMultiple bool) []*evmtypes.MsgEthereumTx {
	//Create evm contract - Owner.sol
	bytecode := ethcommon.FromHex("0x608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16600073ffffffffffffffffffffffffffffffffffffffff167f342827c97908e5e2f71151c08502a66d44b6f758e3ac2f1de95f02eb95f0a73560405160405180910390a36102c4806100dc6000396000f3fe608060405234801561001057600080fd5b5060043610610053576000357c010000000000000000000000000000000000000000000000000000000090048063893d20e814610058578063a6f9dae1146100a2575b600080fd5b6100606100e6565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6100e4600480360360208110156100b857600080fd5b81019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919050505061010f565b005b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146101d1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260138152602001807f43616c6c6572206973206e6f74206f776e65720000000000000000000000000081525060200191505060405180910390fd5b8073ffffffffffffffffffffffffffffffffffffffff166000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff167f342827c97908e5e2f71151c08502a66d44b6f758e3ac2f1de95f02eb95f0a73560405160405180910390a3806000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505056fea265627a7a72315820f397f2733a89198bc7fed0764083694c5b828791f39ebcbc9e414bccef14b48064736f6c63430005100032")
	txs := make([]*evmtypes.MsgEthereumTx, 0)
	for txCount := 0; txCount < totalTxNum/2; txCount++ {
		curTxGp := baseGP + *gpOffset
		if !decreaseGP {
			*gpOffset++
		} else {
			*gpOffset--
		}
		tx := evmtypes.NewMsgEthereumTx(nonce, nil, evmAmountZero, evmGasLimit, big.NewInt(curTxGp), bytecode)

		txs = append(txs, tx)
		nonce++
	}
	multiple := int64(1)
	if needMultiple {
		multiple = 100
	}
	for txCount := totalTxNum / 2; txCount < totalTxNum; txCount++ {
		curTxGp := (baseGP + *gpOffset) * multiple
		if !decreaseGP {
			*gpOffset++
		} else {
			*gpOffset--
		}
		tx := evmtypes.NewMsgEthereumTx(nonce, nil, evmAmountZero, evmGasLimit, big.NewInt(curTxGp), bytecode)

		txs = append(txs, tx)
		nonce++
	}
	return txs
}

func (suite *FakeBlockRecommendGPTestSuite) TestRecommendGP() {
	testCases := []struct {
		title string
		// build txs for one block
		buildTxs func(int, int64, *int64, bool, bool) []*evmtypes.MsgEthereumTx

		gpMaxTxNum   int64
		gpMaxGasUsed int64
		gpMode       int

		expectedTotalGU     []int64
		expectedRecommendGp []string
		blocks              int

		needMultiple bool
		gpDecrease   bool
		tpb          []int
	}{
		{
			title:               "5/5 empty block, higher gp mode",
			buildTxs:            generateEvmTxs,
			gpMaxTxNum:          300,
			gpMaxGasUsed:        1000000,
			gpMode:              0,
			expectedTotalGU:     []int64{0, 0, 0, 0, 0},
			expectedRecommendGp: []string{"100000000", "100000000", "100000000", "100000000", "100000000"},
			blocks:              5,
			needMultiple:        false,
			tpb:                 []int{0, 0, 0, 0, 0},
		},
		{
			title:               "4/6 empty block, gp increase, higher gp mode",
			buildTxs:            generateEvmTxs,
			gpMaxTxNum:          300,
			gpMaxGasUsed:        40000000,
			gpMode:              0,
			expectedTotalGU:     []int64{46329800, 0, 0, 0, 0, 46329800},
			expectedRecommendGp: []string{"100200099", "100000000", "100000000", "100000000", "100000000", "100200099"},
			blocks:              6,
			needMultiple:        false,
			tpb:                 []int{200, 0, 0, 0, 0, 200},
		},
		{
			title:               "4/6 uncongested block, gp increase, higher gp mode",
			buildTxs:            generateEvmTxs,
			gpMaxTxNum:          300,
			gpMaxGasUsed:        40000000,
			gpMode:              0,
			expectedTotalGU:     []int64{46329800, 23164900, 23164900, 23164900, 23164900, 46329800},
			expectedRecommendGp: []string{"100200099", "100000000", "100000000", "100000000", "100000000", "100200500"},
			blocks:              6,
			needMultiple:        false,
			tpb:                 []int{200, 100, 100, 100, 100, 200},
		},
		{
			title:               "2/5 empty block, congestion, gp increase, higher gp mode",
			buildTxs:            generateEvmTxs,
			gpMaxTxNum:          300,
			gpMaxGasUsed:        1000000,
			gpMode:              0,
			expectedTotalGU:     []int64{46329800, 0, 46329800, 0, 46329800},
			expectedRecommendGp: []string{"100200099", "100000000", "100200099", "100000000", "100200299"},
			blocks:              5,
			needMultiple:        false,
			tpb:                 []int{200, 0, 200, 0, 200},
		},
		{
			title:               "congestion, gp increase, higher gp mode",
			buildTxs:            generateEvmTxs,
			gpMaxTxNum:          300,
			gpMaxGasUsed:        1000000,
			gpMode:              0,
			expectedTotalGU:     []int64{46329800, 46329800, 46329800, 46329800, 46329800},
			expectedRecommendGp: []string{"100200099", "100200099", "100200299", "100200499", "100200699"},
			blocks:              5,
			needMultiple:        false,
			tpb:                 []int{200, 200, 200, 200, 200},
		},
		{
			title:               "2/5 empty block, uncongestion, gp increase, higher gp mode",
			buildTxs:            generateEvmTxs,
			gpMaxTxNum:          300,
			gpMaxGasUsed:        60000000,
			gpMode:              0,
			expectedTotalGU:     []int64{46329800, 0, 46329800, 0, 46329800},
			expectedRecommendGp: []string{"100000000", "100000000", "100000000", "100000000", "100000000"},
			blocks:              5,
			needMultiple:        false,
			tpb:                 []int{200, 0, 200, 0, 200},
		},
		{
			title:               "congestion, gp decrease, normal gp mode",
			buildTxs:            generateEvmTxs,
			gpMaxTxNum:          300,
			gpMaxGasUsed:        1000000,
			gpMode:              1,
			expectedTotalGU:     []int64{46329800, 46329800, 46329800, 46329800, 46329800},
			expectedRecommendGp: []string{"100199802", "100199802", "100199801", "100199603", "100199603"},
			blocks:              5,
			needMultiple:        false,
			gpDecrease:          true,
			tpb:                 []int{200, 200, 200, 200, 200},
		},
		{
			title:               "congestion, gp decrease, higher gp mode",
			buildTxs:            generateEvmTxs,
			gpMaxTxNum:          300,
			gpMaxGasUsed:        1000000,
			gpMode:              0,
			expectedTotalGU:     []int64{46329800, 46329800, 46329800, 46329800, 46329800},
			expectedRecommendGp: []string{"100199900", "100199700", "100199700", "100199700", "100199700"},
			blocks:              5,
			needMultiple:        false,
			gpDecrease:          true,
			tpb:                 []int{200, 200, 200, 200, 200},
		},
		{
			title:               "congestion, gp increase, normal mode",
			buildTxs:            generateEvmTxs,
			gpMaxTxNum:          300,
			gpMaxGasUsed:        1000000,
			gpMode:              1,
			expectedTotalGU:     []int64{46329800, 46329800, 46329800, 46329800, 46329800},
			expectedRecommendGp: []string{"100200001", "100200201", "100200400", "100200402", "100200602"},
			blocks:              5,
			needMultiple:        false,
			tpb:                 []int{200, 200, 200, 200, 200},
		},
		{
			title:               "no congestion, gp increase, higher gp mode",
			buildTxs:            generateEvmTxs,
			gpMaxTxNum:          300,
			gpMaxGasUsed:        60000000,
			gpMode:              0,
			expectedTotalGU:     []int64{46329800, 46329800, 46329800, 46329800, 46329800},
			expectedRecommendGp: []string{"100000000", "100000000", "100000000", "100000000", "100000000"},
			blocks:              5,
			needMultiple:        false,
			tpb:                 []int{200, 200, 200, 200, 200},
		},
		{
			title:               "no congestion, gp increase, gp multiple, higher gp mode",
			buildTxs:            generateEvmTxs,
			gpMaxTxNum:          300,
			gpMaxGasUsed:        60000000,
			gpMode:              0,
			expectedTotalGU:     []int64{46329800, 46329800, 46329800, 46329800, 46329800},
			expectedRecommendGp: []string{"100000000", "100000000", "100000000", "100000000", "100000000"},
			blocks:              5,
			needMultiple:        true,
			tpb:                 []int{200, 200, 200, 200, 200},
		},
		{
			title:               "congestion, gp increase, gp multiple, higher gp mode",
			buildTxs:            generateEvmTxs,
			gpMaxTxNum:          300,
			gpMaxGasUsed:        1000000,
			gpMode:              0,
			expectedTotalGU:     []int64{46329800, 46329800, 46329800, 46329800, 46329800},
			expectedRecommendGp: []string{"5060109900", "5060109900", "5060120000", "5060130100", "5060140200"},
			blocks:              5,
			needMultiple:        true,
			tpb:                 []int{200, 200, 200, 200, 200},
		},
		{
			title:               "congestion, gp decrease, gp multiple, higher gp mode",
			buildTxs:            generateEvmTxs,
			gpMaxTxNum:          300,
			gpMaxGasUsed:        1000000,
			gpMode:              0,
			expectedTotalGU:     []int64{46329800, 46329800, 46329800, 46329800, 46329800},
			expectedRecommendGp: []string{"5060094901", "5060084801", "5060084801", "5060084801", "5060084801"},
			blocks:              5,
			needMultiple:        true,
			gpDecrease:          true,
			tpb:                 []int{200, 200, 200, 200, 200},
		},
		{
			title:               "congestion, gp increase, gp multiple, normal mode",
			buildTxs:            generateEvmTxs,
			gpMaxTxNum:          300,
			gpMaxGasUsed:        1000000,
			gpMode:              1,
			expectedTotalGU:     []int64{46329800, 46329800, 46329800, 46329800, 46329800},
			expectedRecommendGp: []string{"100200001", "100200201", "100200400", "100200402", "100200602"},
			blocks:              5,
			needMultiple:        true,
			tpb:                 []int{200, 200, 200, 200, 200},
		},
	}

	appconfig.GetFecConfig().SetDynamicGpWeight(80)
	appconfig.GetFecConfig().SetDynamicGpCheckBlocks(5)
	suite.SetupTest()
	for _, tc := range testCases {

		appconfig.GetFecConfig().SetDynamicGpMaxTxNum(tc.gpMaxTxNum)
		appconfig.GetFecConfig().SetDynamicGpMaxGasUsed(tc.gpMaxGasUsed)
		appconfig.GetFecConfig().SetDynamicGpMode(tc.gpMode)

		// tx serial
		gpOffset := int64(200000)
		baseGP := int64(params.GWei / 10)
		height := int64(2)
		for i := 0; i < tc.blocks; i++ {
			totalGasUsed := int64(0)
			suite.beginFakeBlock(height)
			suite.Run(tc.title+", tx serial", func() {
				txs := tc.buildTxs(tc.tpb[i], baseGP, &gpOffset, tc.gpDecrease, tc.needMultiple)
				for _, tx := range txs {
					tx.Sign(evmChainID, suite.evmSenderPrivKey.ToECDSA())
					txBytes, err := authclient.GetTxEncoder(nil, authclient.WithEthereumTx())(tx)
					suite.Require().NoError(err)
					txReal := suite.app.PreDeliverRealTx(txBytes)
					suite.Require().NotNil(txReal)
					resp := suite.app.DeliverRealTx(txReal)
					totalGasUsed += resp.GasUsed
				}
			})
			//fmt.Println("totalGasUsed: ", totalGasUsed)
			suite.Require().True(totalGasUsed == tc.expectedTotalGU[i], "block gas expect %d, but %d ", tc.expectedTotalGU[i], totalGasUsed)
			suite.endFakeBlock(tc.expectedRecommendGp[i])
			height++
		}
		for !suite.app.gpo.BlockGPQueue.IsEmpty() {
			suite.app.gpo.BlockGPQueue.Pop()
		}

		// tx parallel
		gpOffset = int64(200000)
		baseGP = int64(params.GWei / 10)
		height = int64(2)
		for i := 0; i < tc.blocks; i++ {
			totalGasUsed := int64(0)
			suite.beginFakeBlock(height)
			suite.Run(tc.title+", tx parallel", func() {
				txs := tc.buildTxs(tc.tpb[i], baseGP, &gpOffset, tc.gpDecrease, tc.needMultiple)
				txsBytes := make([][]byte, 0)
				for _, tx := range txs {
					tx.Sign(evmChainID, suite.evmSenderPrivKey.ToECDSA())
					txBytes, err := authclient.GetTxEncoder(nil, authclient.WithEthereumTx())(tx)
					suite.Require().NoError(err)
					txsBytes = append(txsBytes, txBytes)
				}
				resps := suite.app.ParallelTxs(txsBytes, false)
				for _, resp := range resps {
					totalGasUsed += resp.GasUsed
				}
			})
			//fmt.Println("totalGasUsed: ", totalGasUsed)
			suite.Require().True(totalGasUsed == tc.expectedTotalGU[i], "block gas expect %d, but %d ", tc.expectedTotalGU[i], totalGasUsed)
			suite.endFakeBlock(tc.expectedRecommendGp[i])
			height++
		}
		for !suite.app.gpo.BlockGPQueue.IsEmpty() {
			suite.app.gpo.BlockGPQueue.Pop()
		}
	}
}

func TestFakeBlockRecommendGPSuite(t *testing.T) {
	suite.Run(t, new(FakeBlockRecommendGPTestSuite))
}
