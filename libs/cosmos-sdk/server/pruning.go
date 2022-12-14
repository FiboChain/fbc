package server

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/store"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/store/types"
	tmiavl "github.com/FiboChain/fbc/libs/iavl"
)

// GetPruningOptionsFromFlags parses command flags and returns the correct
// PruningOptions. If a pruning strategy is provided, that will be parsed and
// returned, otherwise, it is assumed custom pruning options are provided.
func GetPruningOptionsFromFlags() (types.PruningOptions, error) {
	strategy := strings.ToLower(viper.GetString(FlagPruning))

	switch strategy {
	case types.PruningOptionDefault, types.PruningOptionNothing, types.PruningOptionEverything:
		if strategy == types.PruningOptionNothing {
			tmiavl.EnablePruningHistoryState = false
			tmiavl.CommitIntervalHeight = 1
		}
		return types.NewPruningOptionsFromString(strategy), nil

	case types.PruningOptionCustom:
		opts := types.NewPruningOptions(
			viper.GetUint64(FlagPruningKeepRecent),
			viper.GetUint64(FlagPruningKeepEvery), viper.GetUint64(FlagPruningInterval),
			viper.GetUint64(FlagPruningMaxWsNum),
		)

		if err := opts.Validate(); err != nil {
			return opts, fmt.Errorf("invalid custom pruning options: %w", err)
		}

		return opts, nil

	default:
		return store.PruningOptions{}, fmt.Errorf("unknown pruning strategy %s", strategy)
	}
}
