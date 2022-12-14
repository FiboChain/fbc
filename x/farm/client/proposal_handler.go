package client

import (
	"github.com/FiboChain/fbc/x/farm/client/cli"
	"github.com/FiboChain/fbc/x/farm/client/rest"
	govcli "github.com/FiboChain/fbc/x/gov/client"
)

var (
	// ManageWhiteListProposalHandler alias gov NewProposalHandler
	ManageWhiteListProposalHandler = govcli.NewProposalHandler(cli.GetCmdManageWhiteListProposal, rest.ManageWhiteListProposalRESTHandler)
)
