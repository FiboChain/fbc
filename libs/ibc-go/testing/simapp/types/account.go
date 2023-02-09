package types

import (
	"time"

	cryptotypes "github.com/FiboChain/fbc/libs/cosmos-sdk/crypto/types"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
)

type AccountI interface {
	Copy() AccountI
	GetAddress() sdk.AccAddress
	SetAddress(sdk.AccAddress) error
	GetPubKey() cryptotypes.PubKey
	SetPubKey(cryptotypes.PubKey) error
	GetAccountNumber() uint64
	SetAccountNumber(uint64) error
	GetSequence() uint64
	SetSequence(uint64) error
	GetCoins() sdk.Coins
	SetCoins(sdk.Coins) error
	SpendableCoins(blockTime time.Time) sdk.Coins
	String() string
}
