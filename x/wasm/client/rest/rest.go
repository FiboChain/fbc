package rest

import (
	clientCtx "github.com/FiboChain/fbc/libs/cosmos-sdk/client/context"
	"github.com/gorilla/mux"
)

// RegisterRoutes registers staking-related REST handlers to a router
func RegisterRoutes(cliCtx clientCtx.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
	registerNewTxRoutes(cliCtx, r)
}
