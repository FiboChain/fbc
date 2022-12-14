package ammswap

import (
	"fmt"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/x/ammswap/types"
	tokentypes "github.com/FiboChain/fbc/x/token/types"
)

// GenesisState stores genesis data, all slashing state that must be provided at genesis
type GenesisState struct {
	Params               Params          `json:"params"`
	SwapTokenPairRecords []SwapTokenPair `json:"swap_token_pair_records"`
}

// nolint
func NewGenesisState(swapTokenPairRecords []SwapTokenPair) GenesisState {
	return GenesisState{SwapTokenPairRecords: nil}
}

// ValidateGenesis validates the format of the specified genesisState
func ValidateGenesis(data GenesisState) error {
	for _, record := range data.SwapTokenPairRecords {
		if !record.QuotePooledCoin.IsValid() {
			return fmt.Errorf("invalid SwapTokenPairRecord: QuotePooledCoin: %s", record.QuotePooledCoin.String())
		}
		if !record.BasePooledCoin.IsValid() {
			return fmt.Errorf("invalid SwapTokenPairRecord: BasePooledCoin: %s", record.BasePooledCoin)
		}
		if !tokentypes.NotAllowedOriginSymbol(record.PoolTokenName) {
			return fmt.Errorf("invalid SwapTokenPairRecord: PoolToken: %s. Error: invalid PoolToken", record.PoolTokenName)
		}
	}
	return nil
}

// nolint
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:               types.DefaultParams(),
		SwapTokenPairRecords: nil,
	}
}

// InitGenesis init genesis data to keeper
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetParams(ctx, data.Params)
	for _, record := range data.SwapTokenPairRecords {
		keeper.SetSwapTokenPair(ctx, record.TokenPairName(), record)
	}
}

// ExportGenesis exports genesis from keeper
func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	var records []SwapTokenPair
	iterator := k.GetSwapTokenPairsIterator(ctx)
	for ; iterator.Valid(); iterator.Next() {
		tokenPair := SwapTokenPair{}
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &tokenPair)
		records = append(records, tokenPair)

	}
	params := k.GetParams(ctx)
	return GenesisState{SwapTokenPairRecords: records, Params: params}
}
