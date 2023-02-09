package types

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/tendermint/global"
	"github.com/FiboChain/fbc/libs/tendermint/types"
)

// ensure Msg interface compliance at compile time
var (
	_ sdk.Msg = &MsgEditValidatorCommissionRate{}
)

// MsgEditValidatorCommissionRate - struct for editing a validator commission rate
type MsgEditValidatorCommissionRate struct {
	CommissionRate   sdk.Dec        `json:"commission_rate" yaml:"commission_rate"`
	ValidatorAddress sdk.ValAddress `json:"address" yaml:"address"`
}

// NewMsgEditValidatorCommissionRate creates a msg of edit-validator-commission-rate
func NewMsgEditValidatorCommissionRate(valAddr sdk.ValAddress, newRate sdk.Dec) MsgEditValidatorCommissionRate {
	return MsgEditValidatorCommissionRate{
		CommissionRate:   newRate,
		ValidatorAddress: valAddr,
	}
}

// nolint
func (msg MsgEditValidatorCommissionRate) Route() string { return RouterKey }
func (msg MsgEditValidatorCommissionRate) Type() string  { return "edit_validator_commission_rate" }
func (msg MsgEditValidatorCommissionRate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.ValidatorAddress)}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg MsgEditValidatorCommissionRate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic gives a quick validity check
func (msg MsgEditValidatorCommissionRate) ValidateBasic() error {
	if msg.ValidatorAddress.Empty() {
		return ErrNilValidatorAddr()
	}

	if msg.CommissionRate.GT(sdk.OneDec()) || msg.CommissionRate.IsNegative() {
		return ErrInvalidCommissionRate()
	}

	//will delete it after upgrade venus2
	if !types.HigherThanVenus2(global.GetGlobalHeight()) {
		return ErrCodeNotSupportEditValidatorCommissionRate()
	}
	return nil
}
