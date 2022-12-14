package keeper

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/x/evm/types"
)

// GetParams returns the total set of evm parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	if ctx.IsDeliver() {
		if types.GetEvmParamsCache().IsNeedParamsUpdate() {
			k.paramSpace.GetParamSet(ctx, &params)
			types.GetEvmParamsCache().UpdateParams(params)
		} else {
			params = types.GetEvmParamsCache().GetParams()
		}
	} else {
		k.paramSpace.GetParamSet(ctx, &params)
	}

	return
}

// SetParams sets the evm parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
	types.GetEvmParamsCache().SetNeedParamsUpdate()
}
