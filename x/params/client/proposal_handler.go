package client

import (
	govclient "github.com/FiboChain/fbc/x/gov/client"
	"github.com/FiboChain/fbc/x/params/client/cli"
	"github.com/FiboChain/fbc/x/params/client/rest"
)

// ProposalHandler is the param change proposal handler in cmsdk
var ProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitProposal, rest.ProposalRESTHandler)
