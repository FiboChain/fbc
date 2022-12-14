package client

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/mint/client/cli"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/mint/client/rest"
	govcli "github.com/FiboChain/fbc/x/gov/client"
)

var (
	ManageTreasuresProposalHandler = govcli.NewProposalHandler(
		cli.GetCmdManageTreasuresProposal,
		rest.ManageTreasuresProposalRESTHandler,
	)
)
