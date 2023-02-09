package feesplit

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	sdkerrors "github.com/FiboChain/fbc/libs/cosmos-sdk/types/errors"
	"github.com/FiboChain/fbc/x/common"
	"github.com/FiboChain/fbc/x/feesplit/types"
	govTypes "github.com/FiboChain/fbc/x/gov/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

// NewProposalHandler handles "gov" type message in "feesplit"
func NewProposalHandler(k *Keeper) govTypes.Handler {
	return func(ctx sdk.Context, proposal *govTypes.Proposal) (err sdk.Error) {
		switch content := proposal.Content.(type) {
		case types.FeeSplitSharesProposal:
			return handleFeeSplitSharesProposal(ctx, k, content)
		default:
			return common.ErrUnknownProposalType(types.DefaultCodespace, content.ProposalType())
		}
	}
}

func handleFeeSplitSharesProposal(ctx sdk.Context, k *Keeper, p types.FeeSplitSharesProposal) sdk.Error {
	for _, share := range p.Shares {
		contract := ethcommon.HexToAddress(share.ContractAddr)
		_, found := k.GetFeeSplit(ctx, contract)
		if !found {
			return sdkerrors.Wrapf(
				types.ErrFeeSplitContractNotRegistered,
				"contract %s is not registered", share.ContractAddr,
			)
		}

		k.SetContractShare(ctx, contract, share.Share)
	}
	return nil
}
