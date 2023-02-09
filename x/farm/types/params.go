package types

import (
	"fmt"
	"github.com/FiboChain/fbc/x/common"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/x/params"
)

// Default parameter namespace
const (
	DefaultParamspace        = ModuleName
	defaultQuoteSymbol       = "usdf"
	defaultCreatePoolFee     = "0"
	defaultCreatePoolDeposit = "10"
)

// Parameter store keys
var (
	KeyQuoteSymbol       = []byte("QuoteSymbol")
	KeyCreatePoolFee     = []byte("CreatePoolFee")
	KeyCreatePoolDeposit = []byte("CreatePoolDeposit")
	keyYieldNativeToken  = []byte("YieldNativeToken")
)

// ParamKeyTable for farm module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// Params - used for initializing default parameter for farm at genesis
type Params struct {
	QuoteSymbol       string      `json:"quote_symbol"`
	CreatePoolFee     sdk.SysCoin `json:"create_pool_fee"`
	CreatePoolDeposit sdk.SysCoin `json:"create_pool_deposit"`
	// proposal params
	YieldNativeToken bool `json:"yield_native_token"`
}

// String implements the stringer interface for Params
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  Quote Symbol:								%s
  Create Pool Fee:							%s
  Create Pool Deposit:						%s
  Yield Native Token Enabled:               %v`,
		p.QuoteSymbol, p.CreatePoolFee, p.CreatePoolDeposit, p.YieldNativeToken)
}

// ParamSetPairs - Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: KeyQuoteSymbol, Value: &p.QuoteSymbol, ValidatorFn: common.ValidateDenom("quote symbol")},
		{Key: KeyCreatePoolFee, Value: &p.CreatePoolFee, ValidatorFn: common.ValidateSysCoin("create pool fee")},
		{Key: KeyCreatePoolDeposit, Value: &p.CreatePoolDeposit, ValidatorFn: common.ValidateSysCoin("create pool deposit")},
		{Key: keyYieldNativeToken, Value: &p.YieldNativeToken, ValidatorFn: common.ValidateBool("yield native token")},
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return Params{
		QuoteSymbol:       defaultQuoteSymbol,
		CreatePoolFee:     sdk.NewDecCoinFromDec(common.NativeToken, sdk.MustNewDecFromStr(defaultCreatePoolFee)),
		CreatePoolDeposit: sdk.NewDecCoinFromDec(common.NativeToken, sdk.MustNewDecFromStr(defaultCreatePoolDeposit)),
		YieldNativeToken:  false,
	}
}
