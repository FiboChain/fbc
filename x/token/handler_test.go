package token_test

import (
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	fbchain "github.com/FiboChain/fbc/app"
	app "github.com/FiboChain/fbc/app/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/mock"
	abci "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	"github.com/FiboChain/fbc/libs/tendermint/crypto/secp256k1"
	"github.com/FiboChain/fbc/libs/tendermint/libs/log"
	"github.com/FiboChain/fbc/x/common"
	"github.com/FiboChain/fbc/x/common/version"
	"github.com/FiboChain/fbc/x/token"
	"github.com/FiboChain/fbc/x/token/types"
	"github.com/stretchr/testify/require"
	dbm "github.com/FiboChain/fbc/libs/tm-db"
)

func TestHandlerBlockedContractAddrSend(t *testing.T) {
	fbapp := initApp(true)
	ctx := fbapp.BaseApp.NewContext(true, abci.Header{Height: 1})
	gAcc := CreateEthAccounts(3, sdk.SysCoins{
		sdk.NewDecCoinFromDec(common.NativeToken, sdk.NewDec(10000)),
	})
	fbapp.AccountKeeper.SetAccount(ctx, gAcc[0])
	fbapp.AccountKeeper.SetAccount(ctx, gAcc[1])
	gAcc[2].CodeHash = []byte("contract code hash")
	fbapp.AccountKeeper.SetAccount(ctx, gAcc[2])

	// multi send
	multiSendStr := `[{"to":"` + gAcc[1].Address.String() + `","amount":" 10` + common.NativeToken + `"}]`
	transfers, err := types.StrToTransfers(multiSendStr)
	require.Nil(t, err)
	multiSendStr2 := `[{"to":"` + gAcc[2].Address.String() + `","amount":" 10` + common.NativeToken + `"}]`
	transfers2, err := types.StrToTransfers(multiSendStr2)
	require.Nil(t, err)

	successfulSendMsg := types.NewMsgTokenSend(gAcc[0].Address, gAcc[1].Address, sdk.SysCoins{sdk.NewDecCoinFromDec(common.NativeToken, sdk.NewDec(1))})
	sendToContractMsg := types.NewMsgTokenSend(gAcc[0].Address, gAcc[2].Address, sdk.SysCoins{sdk.NewDecCoinFromDec(common.NativeToken, sdk.NewDec(1))})
	successfulMultiSendMsg := types.NewMsgMultiSend(gAcc[0].Address, transfers)
	multiSendToContractMsg := types.NewMsgMultiSend(gAcc[0].Address, transfers2)
	handler := token.NewTokenHandler(fbapp.TokenKeeper, version.CurrentProtocolVersion)
	fbapp.BankKeeper.SetSendEnabled(ctx, true)
	TestSets := []struct {
		description string
		balance     string
		msg         sdk.Msg
		account     app.EthAccount
	}{
		// 0.01fibo as fixed fee in each stdTx
		{"success to send", "9999.000000000000000000fibo", successfulSendMsg, gAcc[0]},
		{"success to multi-send", "9989.000000000000000000fibo", successfulMultiSendMsg, gAcc[0]},
		{"success to send", "9988.000000000000000000fibo", successfulSendMsg, gAcc[0]},
		{"success to multi-send", "9978.000000000000000000fibo", successfulMultiSendMsg, gAcc[0]},
		//{"fail to send to contract", "9978.000000000000000000fibo", failedSendMsg, gAcc[0]},
		//{"fail to multi-send to contract", "9978.000000000000000000fibo", failedMultiSendMsg, gAcc[0]},
		{"fail to send to contract", "9978.000000000000000000fibo", sendToContractMsg, gAcc[0]},
		{"fail to multi-send to contract", "9978.000000000000000000fibo", multiSendToContractMsg, gAcc[0]},
	}
	for i, tt := range TestSets {
		t.Run(tt.description, func(t *testing.T) {
			handler(ctx, TestSets[i].msg)
			acc := fbapp.AccountKeeper.GetAccount(ctx, tt.account.Address)
			acc.GetCoins().String()
			require.Equal(t, acc.GetCoins().String(), tt.balance)
		})
	}
}

// Setup initializes a new FBchainApp. A Nop logger is set in FBchainApp.
func initApp(isCheckTx bool) *fbchain.FBchainApp {
	db := dbm.NewMemDB()
	app := fbchain.NewFBchainApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, 0)

	if !isCheckTx {
		// init chain must be called to stop deliverState from being nil
		genesisState := fbchain.NewDefaultGenesisState()
		stateBytes, err := codec.MarshalJSONIndent(app.Codec(), genesisState)
		if err != nil {
			panic(err)
		}

		// Initialize the chain
		app.InitChain(
			abci.RequestInitChain{
				Validators:    []abci.ValidatorUpdate{},
				AppStateBytes: stateBytes,
			},
		)
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit(abci.RequestCommit{})
	}

	return app
}

func CreateEthAccounts(numAccs int, genCoins sdk.SysCoins) (genAccs []app.EthAccount) {
	for i := 0; i < numAccs; i++ {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey()
		addr := sdk.AccAddress(pubKey.Address())

		ak := mock.NewAddrKeys(addr, pubKey, privKey)
		testAccount := app.EthAccount{
			BaseAccount: &auth.BaseAccount{
				Address: ak.Address,
				Coins:   genCoins,
			},
			CodeHash: ethcrypto.Keccak256(nil),
		}
		genAccs = append(genAccs, testAccount)
	}
	return
}
