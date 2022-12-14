package types

import (
	"fmt"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
)

// used for import / export via genesis json
type PoolHistoricalRewardsRecord struct {
	PoolName string                `json:"pool_name" yaml:"pool_name"`
	Period   uint64                `json:"period" yaml:"period"`
	Rewards  PoolHistoricalRewards `json:"rewards" yaml:"rewards"`
}

// used for import / export via genesis json
type PoolCurrentRewardsRecord struct {
	PoolName string             `json:"pool_name" yaml:"pool_name"`
	Rewards  PoolCurrentRewards `json:"rewards" yaml:"rewards"`
}

// used for import / export via genesis json
type LockInfoRecord struct {
	PoolName    string         `json:"pool_name" yaml:"pool_name"`
	LockAddress sdk.AccAddress `json:"lock_address" yaml:"lock_address"`
	LockInfo    LockInfo       `json:"lock_info" yaml:"lock_info"`
}

// GenesisState - all farm state that must be provided at genesis
type GenesisState struct {
	Pools                 FarmPools                     `json:"pools" yaml:"pools"`
	LockInfos             []LockInfo                    `json:"lock_infos" yaml:"lock_infos"`
	PoolHistoricalRewards []PoolHistoricalRewardsRecord `json:"historical_rewards" yaml:"historical_rewards"`
	PoolCurrentRewards    []PoolCurrentRewardsRecord    `json:"current_rewards" yaml:"current_rewards"`
	WhiteList             PoolNameList                  `json:"pools_yield_native_token" yaml:"pools_yield_native_token"`
	Params                Params                        `json:"params" yaml:"params"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(pools FarmPools, lockInfos []LockInfo, histories []PoolHistoricalRewardsRecord,
	currents []PoolCurrentRewardsRecord, whiteList PoolNameList, params Params,
) GenesisState {
	return GenesisState{
		Pools:                 pools,
		LockInfos:             lockInfos,
		PoolHistoricalRewards: histories,
		PoolCurrentRewards:    currents,
		WhiteList:             whiteList,
		Params:                params,
	}
}

// DefaultGenesisState - default GenesisState used by Cosmos Hub
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Pools:                 FarmPools{},
		LockInfos:             []LockInfo{},
		PoolHistoricalRewards: []PoolHistoricalRewardsRecord{},
		PoolCurrentRewards:    []PoolCurrentRewardsRecord{},
		Params:                DefaultParams(),
	}
}

// ValidateGenesis validates the farm genesis parameters
func ValidateGenesis(data GenesisState) error {
	if len(data.Pools) != len(data.PoolCurrentRewards) {
		return fmt.Errorf("count of pools(%d) is not equal to that of current rewards(%d)",
			len(data.Pools), len(data.PoolCurrentRewards))
	}

	var expectedReferenceCount uint16
	for _, h := range data.PoolHistoricalRewards {
		expectedReferenceCount += h.Rewards.ReferenceCount
	}

	actualReferenceCount := len(data.LockInfos) + len(data.PoolCurrentRewards)
	if actualReferenceCount != int(expectedReferenceCount) {
		return fmt.Errorf("actual reference count(%d) is not equal to expected reference count(%d)",
			actualReferenceCount, expectedReferenceCount)
	}
	return nil
}
