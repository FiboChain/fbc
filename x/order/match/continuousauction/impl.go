package continuousauction

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"

	"github.com/FiboChain/fbc/x/order/keeper"
)

// nolint
type CaEngine struct {
}

// nolint
func (e *CaEngine) Run(ctx sdk.Context, keeper keeper.Keeper) {
	// TODO
}
