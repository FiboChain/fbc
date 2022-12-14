package rest

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/client/context"

	"github.com/gorilla/mux"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, handlers []EvidenceRESTHandler) {
	// TODO: Register tx handlers.
}
