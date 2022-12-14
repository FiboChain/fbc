package keeper

import (
	"testing"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/x/dex/types"
	"github.com/stretchr/testify/require"
)

func TestModuleAccountInvariant(t *testing.T) {

	testInput := createTestInputWithBalance(t, 1, 10000)
	ctx := testInput.Ctx
	keeper := testInput.DexKeeper
	accounts := testInput.TestAddrs
	keeper.SetParams(ctx, *types.DefaultParams())

	builtInTP := GetBuiltInTokenPair()
	builtInTP.Owner = accounts[0]
	err := keeper.SaveTokenPair(ctx, builtInTP)
	require.Nil(t, err)

	// deposit xxb_fibo 100 fibo
	depositMsg := types.NewMsgDeposit(builtInTP.Name(),
		sdk.NewDecCoin(builtInTP.QuoteAssetSymbol, sdk.NewInt(100)), accounts[0])

	err = keeper.Deposit(ctx, builtInTP.Name(), depositMsg.Depositor, depositMsg.Amount)
	require.Nil(t, err)

	// module acount balance 100fibo
	// xxb_fibo deposits 100 fibo. withdraw info 0 fibo
	invariant := ModuleAccountInvariant(keeper, keeper.supplyKeeper)
	_, broken := invariant(ctx)
	require.False(t, broken)

	// withdraw xxb_fibo 50 fibo
	WithdrawMsg := types.NewMsgWithdraw(builtInTP.Name(),
		sdk.NewDecCoin(builtInTP.QuoteAssetSymbol, sdk.NewInt(50)), accounts[0])

	err = keeper.Withdraw(ctx, builtInTP.Name(), WithdrawMsg.Depositor, WithdrawMsg.Amount)
	require.Nil(t, err)

	// module acount balance 100fibo
	// xxb_fibo deposits 50 fibo. withdraw info 50 fibo
	invariant = ModuleAccountInvariant(keeper, keeper.supplyKeeper)
	_, broken = invariant(ctx)
	require.False(t, broken)

}
