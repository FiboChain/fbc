package transfer

import (
	"github.com/FiboChain/fbc/libs/ibc-go/modules/apps/transfer/keeper"
	"github.com/FiboChain/fbc/libs/ibc-go/modules/apps/transfer/types"
)

var (
	NewKeeper  = keeper.NewKeeper
	ModuleCdc  = types.ModuleCdc
	SetMarshal = types.SetMarshal
)
