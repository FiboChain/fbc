package v018

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/x/evm/legacy/v0_16"
	evmTypes "github.com/FiboChain/fbc/x/evm/types"
)

const (
	ModuleName = "evm"
)

// Migrate adds contract
func Migrate(oldGenState v0_16.GenesisState) GenesisState {
	params := Params{
		EnableCreate:                      false,
		EnableCall:                        false,
		ExtraEIPs:                         oldGenState.Params.ExtraEIPs,
		EnableContractDeploymentWhitelist: true,
		EnableContractBlockedList:         true,
		MaxGasLimitPerTx:                  evmTypes.DefaultMaxGasLimitPerTx,
	}

	return GenesisState{
		Accounts:                    oldGenState.Accounts,
		TxsLogs:                     oldGenState.TxsLogs,
		ContractDeploymentWhitelist: []sdk.AccAddress{},
		ContractBlockedList:         []sdk.AccAddress{},
		ChainConfig:                 oldGenState.ChainConfig,
		Params:                      params,
	}
}
