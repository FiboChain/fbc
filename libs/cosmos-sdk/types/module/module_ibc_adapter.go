package module

import (
	clientCtx "github.com/FiboChain/fbc/libs/cosmos-sdk/client/context"
	codectypes "github.com/FiboChain/fbc/libs/cosmos-sdk/codec/types"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// RegisterInterfaces registers all module interface types
func (bm BasicManager) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	for _, m := range bm {
		if ada, ok := m.(AppModuleBasicAdapter); ok {
			ada.RegisterInterfaces(registry)
		}
	}
}

// RegisterGRPCGatewayRoutes registers all module rest routes
func (bm BasicManager) RegisterGRPCGatewayRoutes(clientCtx clientCtx.CLIContext, rtr *runtime.ServeMux) {
	for _, m := range bm {
		if ada, ok := m.(AppModuleBasicAdapter); ok {
			ada.RegisterGRPCGatewayRoutes(clientCtx, rtr)
		}
	}
}

func (bm BasicManager) RegisterRPCRouterForGRPC(clientCtx clientCtx.CLIContext, rtr *mux.Router) {
	for _, m := range bm {
		if ada, ok := m.(AppModuleBasicAdapter); ok {
			ada.RegisterRouterForGRPC(clientCtx, rtr)
		}
	}
}
