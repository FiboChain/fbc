package connection

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	interfacetypes "github.com/FiboChain/fbc/libs/cosmos-sdk/codec/types"
	"github.com/FiboChain/fbc/libs/ibc-go/modules/core/03-connection/client/cli"
	"github.com/FiboChain/fbc/libs/ibc-go/modules/core/03-connection/types"
	"github.com/gogo/protobuf/grpc"
	"github.com/spf13/cobra"
)

// Name returns the IBC connection ICS name.
func Name() string {
	return types.SubModuleName
}

// GetQueryCmd returns the root query command for the IBC connections.
func GetQueryCmd(cdc *codec.CodecProxy, reg interfacetypes.InterfaceRegistry) *cobra.Command {
	return cli.GetQueryCmd(cdc, reg)
}

// RegisterQueryService registers the gRPC query service for IBC connections.
func RegisterQueryService(server grpc.Server, queryServer types.QueryServer) {
	types.RegisterQueryServer(server, queryServer)
}
