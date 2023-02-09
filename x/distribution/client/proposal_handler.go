package client

import (
	"github.com/FiboChain/fbc/x/distribution/client/cli"
	"github.com/FiboChain/fbc/x/distribution/client/rest"
	govclient "github.com/FiboChain/fbc/x/gov/client"
)

// param change proposal handler
var (
	CommunityPoolSpendProposalHandler      = govclient.NewProposalHandler(cli.GetCmdCommunityPoolSpendProposal, rest.CommunityPoolSpendProposalRESTHandler)
	ChangeDistributionTypeProposalHandler  = govclient.NewProposalHandler(cli.GetChangeDistributionTypeProposal, rest.ChangeDistributionTypeProposalRESTHandler)
	WithdrawRewardEnabledProposalHandler   = govclient.NewProposalHandler(cli.GetWithdrawRewardEnabledProposal, rest.WithdrawRewardEnabledProposalRESTHandler)
	RewardTruncatePrecisionProposalHandler = govclient.NewProposalHandler(cli.GetRewardTruncatePrecisionProposal, rest.RewardTruncatePrecisionProposalRESTHandler)
)
