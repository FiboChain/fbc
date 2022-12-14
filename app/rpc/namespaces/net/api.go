package net

import (
	"fmt"

	"github.com/FiboChain/fbc/app/rpc/monitor"
	ethermint "github.com/FiboChain/fbc/app/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/client/context"
	"github.com/FiboChain/fbc/libs/tendermint/libs/log"
	rpcclient "github.com/FiboChain/fbc/libs/tendermint/rpc/client"
)

// PublicNetAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicNetAPI struct {
	networkVersion uint64
	logger         log.Logger
	Metrics        map[string]*monitor.RpcMetrics
	tmClient       rpcclient.Client
}

// NewAPI creates an instance of the public Net Web3 API.
func NewAPI(clientCtx context.CLIContext, log log.Logger) *PublicNetAPI {
	// parse the chainID from a integer string
	chainIDEpoch, err := ethermint.ParseChainID(clientCtx.ChainID)
	if err != nil {
		panic(err)
	}

	return &PublicNetAPI{
		networkVersion: chainIDEpoch.Uint64(),
		logger:         log.With("module", "json-rpc", "namespace", "net"),
		tmClient:       clientCtx.Client,
	}
}

// Version returns the current ethereum protocol version.
func (api *PublicNetAPI) Version() string {
	monitor := monitor.GetMonitor("net_version", api.logger, api.Metrics).OnBegin()
	defer monitor.OnEnd()
	return fmt.Sprintf("%d", api.networkVersion)
}

// Listening returns if client is actively listening for network connections.
func (api *PublicNetAPI) Listening() bool {
	monitor := monitor.GetMonitor("net_listening", api.logger, api.Metrics).OnBegin()
	defer monitor.OnEnd()
	netInfo, err := api.tmClient.NetInfo()
	if err != nil {
		return false
	}
	return netInfo.Listening
}

// PeerCount returns the number of peers currently connected to the client.
func (api *PublicNetAPI) PeerCount() int {
	monitor := monitor.GetMonitor("net_peerCount", api.logger, api.Metrics).OnBegin()
	defer monitor.OnEnd()
	netInfo, err := api.tmClient.NetInfo()
	if err != nil {
		return 0
	}
	return len(netInfo.Peers)
}