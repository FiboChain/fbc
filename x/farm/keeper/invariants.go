package keeper

import (
	"fmt"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/x/farm/types"
)

// RegisterInvariants registers all farm invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "module-account", moduleAccountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "yield-farming-account", yieldFarmingAccountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "mint-farming-account", mintFarmingAccountInvariant(k))
}

// moduleAccountInvariant checks if farm ModuleAccount is consistent with the sum of deposit amount
func moduleAccountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// iterate all pools, then calculate the total deposit amount
		totalDepositAmount := sdk.SysCoins{}
		pools := k.GetFarmPools(ctx)
		for _, pool := range pools {
			totalDepositAmount = totalDepositAmount.Add2(pool.DepositAmount.ToCoins())
		}

		// iterate all lock infos
		totalLockedAmount := sdk.SysCoins{}
		k.IterateAllLockInfos(ctx, func(lockInfo types.LockInfo) (stop bool) {
			totalLockedAmount = totalLockedAmount.Add2(sdk.NewDecCoins(lockInfo.Amount))
			return false
		})

		// get farm module account
		moduleAccAmount := k.SupplyKeeper().GetModuleAccount(ctx, types.ModuleName).GetCoins()

		// make a comparison
		broken := !(moduleAccAmount.IsEqual(totalDepositAmount.Add2(totalLockedAmount)))

		return sdk.FormatInvariant(types.ModuleName, "ModuleAccount coins",
			fmt.Sprintf("\texpected farm ModuleAccount coins: %s\n"+
				"\tacutal farm ModuleAccount coins: %s\n",
				totalDepositAmount, moduleAccAmount)), broken
	}
}

// yieldFarmingAccountInvariant checks if yield_farming_account ModuleAccount is consistent
// with the total accumulated rewards
func yieldFarmingAccountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// iterate all pools, then calculate the total deposit amount
		expectedYieldModuleAccAmount := sdk.SysCoins{}
		pools := k.GetFarmPools(ctx)
		for _, pool := range pools {
			expectedYieldModuleAccAmount = expectedYieldModuleAccAmount.Add2(pool.TotalAccumulatedRewards)
			for _, yieldInfo := range pool.YieldedTokenInfos {
				expectedYieldModuleAccAmount = expectedYieldModuleAccAmount.Add2(sdk.SysCoins{yieldInfo.RemainingAmount})
			}
		}

		// get yield_farming_account module account
		actualYieldModuleAccAmount := k.SupplyKeeper().GetModuleAccount(ctx, types.YieldFarmingAccount).GetCoins()

		// make a comparison
		broken := !(expectedYieldModuleAccAmount.IsEqual(actualYieldModuleAccAmount))

		return sdk.FormatInvariant(types.ModuleName, "yield_farming_account coins",
			fmt.Sprintf("\texpected yield_farming_account coins: %s\n"+
				"\tacutal yield_farming_account coins: %s\n",
				expectedYieldModuleAccAmount, actualYieldModuleAccAmount)), broken
	}
}

// mintFarmingAccountInvariant checks if mint_farming_account ModuleAccount is consistent
// with the sum of yielded native tokens
func mintFarmingAccountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		broken := false

		// get mint_farming_account module account
		moduleAcc := k.SupplyKeeper().GetModuleAccount(ctx, types.MintFarmingAccount)

		// get white_lists
		whiteLists := k.GetWhitelist(ctx)
		if len(whiteLists) != 0 {
			if !moduleAcc.GetCoins().IsZero() {
				broken = true
			}
		}

		return sdk.FormatInvariant(types.ModuleName, "mint_farming_account coins",
			fmt.Sprintf("\texpected mint_farming_account coins should be zero\n"+
				"\tacutal mint_farming_account coins: %s\n"+
				"\twhite lists: %s\n",
				moduleAcc.GetCoins(), whiteLists)), broken
	}
}
