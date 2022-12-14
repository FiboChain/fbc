package client

import (
	"github.com/FiboChain/fbc/x/distribution/client/cli"
	"github.com/FiboChain/fbc/x/distribution/client/rest"
	govclient "github.com/FiboChain/fbc/x/gov/client"
)

// param change proposal handler
var (
	ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
)
