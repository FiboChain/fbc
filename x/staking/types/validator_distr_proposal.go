package types

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
)

// ToValAddresses converts []Validators to []sdk.ValAddress
func (v Validators) ToValAddresses() (valAddrs []sdk.ValAddress) {
	for _, val := range v {
		valAddrs = append(valAddrs, val.OperatorAddress)
	}
	return valAddrs
}
