package module

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	"github.com/gogo/protobuf/grpc"
)

// Configurator provides the hooks to allow modules to configure and register
// their services in the RegisterServices method. It is designed to eventually
// support module object capabilities isolation as described in
// https://github.com/cosmos/cosmos-sdk/issues/7093
type Configurator interface {
	// MsgServer returns a grpc.Server instance which allows registering services
	// that will handle TxBody.messages in transactions. These Msg's WILL NOT
	// be exposed as gRPC services.
	MsgServer() grpc.Server

	// QueryServer returns a grpc.Server instance which allows registering services
	// that will be exposed as gRPC services as well as ABCI query handlers.
	QueryServer() grpc.Server
}

type configurator struct {
	cdc         *codec.Codec
	msgServer   grpc.Server
	queryServer grpc.Server
	migrations  map[string]map[uint64]MigrationHandler
}

// NewConfigurator returns a new Configurator instance
func NewConfigurator(cdc *codec.Codec, msgServer grpc.Server, queryServer grpc.Server) Configurator {
	return configurator{
		cdc:         cdc,
		msgServer:   msgServer,
		queryServer: queryServer,
		migrations:  map[string]map[uint64]MigrationHandler{},
	}
}

var _ Configurator = configurator{}

// MsgServer implements the Configurator.MsgServer method
func (c configurator) MsgServer() grpc.Server {
	return c.msgServer
}

// QueryServer implements the Configurator.QueryServer method
func (c configurator) QueryServer() grpc.Server {
	return c.queryServer
}
