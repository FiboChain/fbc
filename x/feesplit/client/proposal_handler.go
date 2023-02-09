package client

import (
	"github.com/FiboChain/fbc/x/feesplit/client/cli"
	"github.com/FiboChain/fbc/x/feesplit/client/rest"
	govcli "github.com/FiboChain/fbc/x/gov/client"
)

var (
	// FeeSplitSharesProposalHandler alias gov NewProposalHandler
	FeeSplitSharesProposalHandler = govcli.NewProposalHandler(
		cli.GetCmdFeeSplitSharesProposal,
		rest.FeeSplitSharesProposalRESTHandler,
	)
)
