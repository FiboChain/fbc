package types

import (
	"time"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
)

// DefaultOwnershipConfirmWindow defines default confirm window
const DefaultOwnershipConfirmWindow = 24 * time.Hour

type ConfirmOwnership struct {
	Product     string         `json:"product"`
	FromAddress sdk.AccAddress `json:"from_address"`
	ToAddress   sdk.AccAddress `json:"to_address"`
	Expire      time.Time      `json:expire`
}
