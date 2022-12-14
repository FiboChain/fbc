package types

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	authexported "github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/exported"
)

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authexported.Account
	IterateAccounts(ctx sdk.Context, cb func(account authexported.Account) bool)
}
