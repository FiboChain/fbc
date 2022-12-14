package client

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/distribution/client/cli"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/distribution/client/rest"
	govclient "github.com/FiboChain/fbc/libs/cosmos-sdk/x/gov/client"
)

// param change proposal handler
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
)
