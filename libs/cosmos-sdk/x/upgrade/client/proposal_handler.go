package client

import (
	govclient "github.com/FiboChain/fbc/libs/cosmos-sdk/x/gov/client"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/upgrade/client/cli"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/upgrade/client/rest"
)

var ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitUpgradeProposal, rest.ProposalRESTHandler)
