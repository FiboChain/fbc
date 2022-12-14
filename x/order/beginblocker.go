package order

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"

	"github.com/FiboChain/fbc/x/common/perf"
	"github.com/FiboChain/fbc/x/order/keeper"
	"github.com/FiboChain/fbc/x/order/types"
	//"github.com/FiboChain/fbc/x/common/version"
)

// BeginBlocker runs the logic of BeginBlocker with version 0.
// BeginBlocker resets keeper cache.
func BeginBlocker(ctx sdk.Context, keeper keeper.Keeper) {
	seq := perf.GetPerf().OnBeginBlockEnter(ctx, types.ModuleName)
	defer perf.GetPerf().OnBeginBlockExit(ctx, types.ModuleName, seq)

	keeper.ResetCache(ctx)
}
