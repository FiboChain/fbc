package main

import (
	"fmt"
	mintclient "github.com/FiboChain/fbc/libs/cosmos-sdk/x/mint/client"
	evmclient "github.com/FiboChain/fbc/x/evm/client"

	"github.com/FiboChain/fbc/app/rpc"
	"github.com/FiboChain/fbc/app/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/client"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/client/lcd"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/server"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth"
	authrest "github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth/client/rest"
	bankrest "github.com/FiboChain/fbc/libs/cosmos-sdk/x/bank/client/rest"
	supplyrest "github.com/FiboChain/fbc/libs/cosmos-sdk/x/supply/client/rest"
	ammswaprest "github.com/FiboChain/fbc/x/ammswap/client/rest"
	dexclient "github.com/FiboChain/fbc/x/dex/client"
	dexrest "github.com/FiboChain/fbc/x/dex/client/rest"
	dist "github.com/FiboChain/fbc/x/distribution"
	distr "github.com/FiboChain/fbc/x/distribution"
	distrest "github.com/FiboChain/fbc/x/distribution/client/rest"
	evmrest "github.com/FiboChain/fbc/x/evm/client/rest"
	farmclient "github.com/FiboChain/fbc/x/farm/client"
	farmrest "github.com/FiboChain/fbc/x/farm/client/rest"
	govrest "github.com/FiboChain/fbc/x/gov/client/rest"
	orderrest "github.com/FiboChain/fbc/x/order/client/rest"
	paramsclient "github.com/FiboChain/fbc/x/params/client"
	stakingrest "github.com/FiboChain/fbc/x/staking/client/rest"
	"github.com/FiboChain/fbc/x/token"
	tokensrest "github.com/FiboChain/fbc/x/token/client/rest"
	"github.com/spf13/viper"
)

// registerRoutes registers the routes from the different modules for the LCD.
// NOTE: details on the routes added for each module are in the module documentation
// NOTE: If making updates here you also need to update the test helper in client/lcd/test_helper.go
func registerRoutes(rs *lcd.RestServer) {
	rpc.RegisterRoutes(rs)
	pathPrefix := viper.GetString(server.FlagRestPathPrefix)
	if pathPrefix == "" {
		pathPrefix = types.EthBech32Prefix
	}
	registerRoutesV1(rs, pathPrefix)
	registerRoutesV2(rs, pathPrefix)
}

func registerRoutesV1(rs *lcd.RestServer, pathPrefix string) {
	v1Router := rs.Mux.PathPrefix(fmt.Sprintf("/%s/v1", pathPrefix)).Name("v1").Subrouter()
	client.RegisterRoutes(rs.CliCtx, v1Router)
	authrest.RegisterRoutes(rs.CliCtx, v1Router, auth.StoreKey)
	bankrest.RegisterRoutes(rs.CliCtx, v1Router)
	stakingrest.RegisterRoutes(rs.CliCtx, v1Router)
	distrest.RegisterRoutes(rs.CliCtx, v1Router, dist.StoreKey)

	orderrest.RegisterRoutes(rs.CliCtx, v1Router)
	tokensrest.RegisterRoutes(rs.CliCtx, v1Router, token.StoreKey)
	dexrest.RegisterRoutes(rs.CliCtx, v1Router)
	ammswaprest.RegisterRoutes(rs.CliCtx, v1Router)
	supplyrest.RegisterRoutes(rs.CliCtx, v1Router)
	farmrest.RegisterRoutes(rs.CliCtx, v1Router)
	evmrest.RegisterRoutes(rs.CliCtx, v1Router)
	govrest.RegisterRoutes(rs.CliCtx, v1Router,
		[]govrest.ProposalRESTHandler{
			paramsclient.ProposalHandler.RESTHandler(rs.CliCtx),
			distr.ProposalHandler.RESTHandler(rs.CliCtx),
			dexclient.DelistProposalHandler.RESTHandler(rs.CliCtx),
			farmclient.ManageWhiteListProposalHandler.RESTHandler(rs.CliCtx),
			evmclient.ManageContractDeploymentWhitelistProposalHandler.RESTHandler(rs.CliCtx),
			mintclient.ManageTreasuresProposalHandler.RESTHandler(rs.CliCtx),
		},
	)
}

func registerRoutesV2(rs *lcd.RestServer, pathPrefix string) {
	v2Router := rs.Mux.PathPrefix(fmt.Sprintf("/%s/v2", pathPrefix)).Name("v1").Subrouter()
	client.RegisterRoutes(rs.CliCtx, v2Router)
	authrest.RegisterRoutes(rs.CliCtx, v2Router, auth.StoreKey)
	bankrest.RegisterRoutes(rs.CliCtx, v2Router)
	stakingrest.RegisterRoutes(rs.CliCtx, v2Router)
	distrest.RegisterRoutes(rs.CliCtx, v2Router, dist.StoreKey)
	orderrest.RegisterRoutesV2(rs.CliCtx, v2Router)
	tokensrest.RegisterRoutesV2(rs.CliCtx, v2Router, token.StoreKey)
}
