package ibc_tx

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
)

type DenomAdapterMsg interface {
	sdk.Msg
	DenomOpr
}

type DenomOpr interface {
	RulesFilter() (sdk.Msg, error)
}

type MessageSensitive interface {
	Swap(ctx sdk.Context) (sdk.Msg, error)
}
