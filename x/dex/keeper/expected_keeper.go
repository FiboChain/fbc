package keeper

import (
	"time"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/supply/exported"
	"github.com/FiboChain/fbc/x/dex/types"
	ordertypes "github.com/FiboChain/fbc/x/order/types"
	"github.com/FiboChain/fbc/x/params"
)

// SupplyKeeper defines the expected supply Keeper
type SupplyKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress,
		recipientModule string, amt sdk.Coins) sdk.Error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string,
		recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
	GetModuleAccount(ctx sdk.Context, moduleName string) exported.ModuleAccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error
}

// TokenKeeper defines the expected token Keeper
type TokenKeeper interface {
	TokenExist(ctx sdk.Context, symbol string) bool
}

// IKeeper defines the expected dex Keeper
type IKeeper interface {
	GetTokenPair(ctx sdk.Context, product string) *types.TokenPair
	GetTokenPairs(ctx sdk.Context) []*types.TokenPair
	GetUserTokenPairs(ctx sdk.Context, owner sdk.AccAddress) []*types.TokenPair
	GetTokenPairsOrdered(ctx sdk.Context) types.TokenPairs
	SaveTokenPair(ctx sdk.Context, tokenPair *types.TokenPair) error
	DeleteTokenPairByName(ctx sdk.Context, owner sdk.AccAddress, tokenPairName string)
	Deposit(ctx sdk.Context, product string, from sdk.AccAddress, amount sdk.SysCoin) sdk.Error
	Withdraw(ctx sdk.Context, product string, to sdk.AccAddress, amount sdk.SysCoin) sdk.Error
	GetSupplyKeeper() SupplyKeeper
	GetTokenKeeper() TokenKeeper
	GetBankKeeper() BankKeeper
	GetParamSubspace() params.Subspace
	GetParams(ctx sdk.Context) (params types.Params)
	SetParams(ctx sdk.Context, params types.Params)
	GetFeeCollector() string
	LockTokenPair(ctx sdk.Context, product string, lock *ordertypes.ProductLock)
	LoadProductLocks(ctx sdk.Context) *ordertypes.ProductLockMap
	SetWithdrawInfo(ctx sdk.Context, withdrawInfo types.WithdrawInfo)
	SetWithdrawCompleteTimeAddress(ctx sdk.Context, completeTime time.Time, addr sdk.AccAddress)
	IterateWithdrawAddress(ctx sdk.Context, currentTime time.Time, fn func(index int64, key []byte) (stop bool))
	CompleteWithdraw(ctx sdk.Context, addr sdk.AccAddress) error
	IterateWithdrawInfo(ctx sdk.Context, fn func(index int64, withdrawInfo types.WithdrawInfo) (stop bool))
	DeleteWithdrawCompleteTimeAddress(ctx sdk.Context, timestamp time.Time, delAddr sdk.AccAddress)
	SetOperator(ctx sdk.Context, operator types.DEXOperator)
	GetOperator(ctx sdk.Context, addr sdk.AccAddress) (operator types.DEXOperator, isExist bool)
	IterateOperators(ctx sdk.Context, cb func(operator types.DEXOperator) (stop bool))
	GetMaxTokenPairID(ctx sdk.Context) (tokenPairMaxID uint64)
	SetMaxTokenPairID(ctx sdk.Context, tokenPairMaxID uint64)
	GetConfirmOwnership(ctx sdk.Context, product string) (confirmOwnership *types.ConfirmOwnership, exist bool)
	SetConfirmOwnership(ctx sdk.Context, confirmOwnership *types.ConfirmOwnership)
	DeleteConfirmOwnership(ctx sdk.Context, product string)
	UpdateUserTokenPair(ctx sdk.Context, product string, owner, to sdk.AccAddress)
	UpdateTokenPair(ctx sdk.Context, product string, tokenPair *types.TokenPair)
}

// StakingKeeper defines the expected staking Keeper (noalias)
type StakingKeeper interface {
	IsValidator(ctx sdk.Context, addr sdk.AccAddress) bool
}

// BankKeeper defines the expected bank Keeper
type BankKeeper interface {
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// GovKeeper defines the expected gov Keeper
type GovKeeper interface {
	RemoveFromActiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time)
}

type StreamKeeper interface {
	OnAddNewTokenPair(ctx sdk.Context, tokenPair *types.TokenPair)
	OnTokenPairUpdated(ctx sdk.Context)
}
