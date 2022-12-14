package types

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/supply"
	token "github.com/FiboChain/fbc/x/token/types"

	"fmt"
	"strings"
)

// PoolTokenPrefix defines pool token prefix name
const PoolTokenPrefix = "ammswap_"

// SwapTokenPair defines token pair exchange
type SwapTokenPair struct {
	QuotePooledCoin sdk.SysCoin `json:"quote_pooled_coin"` // The volume of quote token in the token pair exchange pool
	BasePooledCoin  sdk.SysCoin `json:"base_pooled_coin"`  // The volume of base token in the token pair exchange pool
	PoolTokenName   string      `json:"pool_token_name"`   // The name of pool token
}

func NewSwapPair(token0, token1 string) SwapTokenPair {
	base, quote := GetBaseQuoteTokenName(token0, token1)

	swapTokenPair := SwapTokenPair{
		sdk.NewDecCoinFromDec(quote, sdk.ZeroDec()),
		sdk.NewDecCoinFromDec(base, sdk.ZeroDec()),
		GetPoolTokenName(token0, token1),
	}
	return swapTokenPair
}

// NewSwapTokenPair is a constructor function for SwapTokenPair
func NewSwapTokenPair(quotePooledCoin sdk.SysCoin, basePooledCoin sdk.SysCoin, poolTokenName string) *SwapTokenPair {
	swapTokenPair := &SwapTokenPair{
		QuotePooledCoin: quotePooledCoin,
		BasePooledCoin:  basePooledCoin,
		PoolTokenName:   poolTokenName,
	}
	return swapTokenPair
}

// String implement fmt.Stringer
func (s SwapTokenPair) String() string {
	return strings.TrimSpace(fmt.Sprintf(`QuotePooledCoin: %s
BasePooledCoin: %s
PoolTokenName: %s`, s.QuotePooledCoin.String(), s.BasePooledCoin.String(), s.PoolTokenName))
}

// TokenPairName defines token pair
func (s SwapTokenPair) TokenPairName() string {
	return s.BasePooledCoin.Denom + "_" + s.QuotePooledCoin.Denom
}

// InitPoolToken default pool token
func InitPoolToken(poolTokenName string) token.Token {
	return token.Token{
		Description:         poolTokenName,
		Symbol:              poolTokenName,
		OriginalSymbol:      poolTokenName,
		WholeName:           poolTokenName,
		OriginalTotalSupply: sdk.NewDec(0),
		Owner:               supply.NewModuleAddress(ModuleName),
		Type:                GenerateTokenType,
		Mintable:            true,
	}
}

func GetSwapTokenPairName(token0, token1 string) string {
	baseTokenName, quoteTokenName := GetBaseQuoteTokenName(token0, token1)
	return baseTokenName + "_" + quoteTokenName
}

func GetBaseQuoteTokenName(token0, token1 string) (string, string) {
	if token0 < token1 {
		return token0, token1
	}
	return token1, token0
}

func ValidateBaseAndQuoteAmount(baseAmountName, quoteAmountName string) error {
	if baseAmountName > quoteAmountName {
		return ErrBaseAmountNameBiggerThanQuoteAmountName()
	} else if baseAmountName == quoteAmountName {
		return ErrBaseNameEqualQuoteName()
	}
	if err := ValidateSwapAmountName(baseAmountName); err != nil {
		return err
	}

	if err := ValidateSwapAmountName(quoteAmountName); err != nil {
		return err
	}
	return nil
}

func ValidateSwapAmountName(amountName string) error {
	if sdk.ValidateDenom(amountName) != nil {
		return ErrValidateDenom(amountName)
	}
	if token.NotAllowedOriginSymbol(amountName) {
		return ErrNotAllowedOriginSymbol()
	}
	return nil
}

func GetPoolTokenName(token1, token2 string) string {
	return PoolTokenPrefix + GetSwapTokenPairName(token1, token2)
}

func IsPoolToken(symbol string) bool {
	return token.NotAllowedOriginSymbol(symbol)
}

func SplitPoolToken(symbol string) (token0, token1 string) {
	splits := strings.Split(symbol, "_")[1:]
	token0 = splits[0]
	token1 = splits[1]
	return
}
