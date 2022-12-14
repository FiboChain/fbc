package client

import (
	"github.com/FiboChain/fbc/x/dex/client/cli"
	"github.com/FiboChain/fbc/x/dex/client/rest"
	govclient "github.com/FiboChain/fbc/x/gov/client"
)

// param change proposal handler
var (
	// DelistProposalHandler alias gov NewProposalHandler
	DelistProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitDelistProposal, rest.DelistProposalRESTHandler)
)
