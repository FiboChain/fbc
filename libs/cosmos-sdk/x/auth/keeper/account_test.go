package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/exported"
	abci "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func Test_IterateAccounts(t *testing.T) {
	var cases = []struct {
		num int
	}{
		{0},
		{1},
		{100},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			app, ctx := createTestAppWithHeight(false, 10)

			addrs := make(map[ethcmn.Address]struct{}, c.num)
			for i := 0; i < c.num; i++ {
				arr := []byte{byte((i & 0xFF0000) >> 16), byte((i & 0xFF00) >> 8), byte(i & 0xFF)}
				addr := sdk.AccAddress(arr)
				addrs[ethcmn.BytesToAddress(arr)] = struct{}{}
				acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
				app.AccountKeeper.SetAccount(ctx, acc)
			}
			app.Commit(abci.RequestCommit{})

			count := 0
			iAddrs := make(map[ethcmn.Address]struct{}, c.num)
			app.AccountKeeper.IterateAccounts(ctx, func(acc exported.Account) bool {
				addr := ethcmn.BytesToAddress(acc.GetAddress())
				if _, ok := addrs[addr]; ok {
					iAddrs[addr] = struct{}{}
					count++
				}
				return false
			})
			require.EqualValues(t, addrs, iAddrs)
			require.Equal(t, len(iAddrs), count)
			require.Equal(t, c.num, count)
		})
	}
}
