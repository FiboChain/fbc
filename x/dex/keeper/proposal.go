package keeper

import (
	"fmt"
	"time"

	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/x/common"
	"github.com/FiboChain/fbc/x/dex/types"
	"github.com/FiboChain/fbc/x/gov"
	govTypes "github.com/FiboChain/fbc/x/gov/types"
)

// GetMinDeposit returns min deposit
func (k Keeper) GetMinDeposit(ctx sdk.Context, content gov.Content) (minDeposit sdk.SysCoins) {
	if _, ok := content.(types.DelistProposal); ok {
		minDeposit = k.GetParams(ctx).DelistMinDeposit
	}
	return
}

// GetMaxDepositPeriod returns max deposit period
func (k Keeper) GetMaxDepositPeriod(ctx sdk.Context, content gov.Content) (maxDepositPeriod time.Duration) {
	if _, ok := content.(types.DelistProposal); ok {
		maxDepositPeriod = k.GetParams(ctx).DelistMaxDepositPeriod
	}
	return
}

// GetVotingPeriod returns voting period
func (k Keeper) GetVotingPeriod(ctx sdk.Context, content gov.Content) (votingPeriod time.Duration) {
	if _, ok := content.(types.DelistProposal); ok {
		votingPeriod = k.GetParams(ctx).DelistVotingPeriod
	}
	return
}

// check msg Delist proposal
func (k Keeper) checkMsgDelistProposal(ctx sdk.Context, delistProposal types.DelistProposal, proposer sdk.AccAddress, initialDeposit sdk.SysCoins) sdk.Error {
	// check the proposer of the msg is a validator
	if !k.stakingKeeper.IsValidator(ctx, proposer) {
		return gov.ErrInvalidProposer()
	}

	// check the propose of the msg is equal the proposer in proposal content
	if !proposer.Equals(delistProposal.Proposer) {
		return gov.ErrInvalidProposer()
	}

	// check whether the baseAsset is in the Dex list
	queryTokenPair := k.GetTokenPair(ctx, fmt.Sprintf("%s_%s", delistProposal.BaseAsset, delistProposal.QuoteAsset))
	if queryTokenPair == nil {
		return types.ErrTokenPairNotFound(fmt.Sprintf("%s_%s", delistProposal.BaseAsset, delistProposal.QuoteAsset))
	}

	// check the initial deposit
	localMinDeposit := k.GetParams(ctx).DelistMinDeposit.MulDec(sdk.NewDecWithPrec(1, 1))
	err := common.HasSufficientCoins(proposer, initialDeposit, localMinDeposit)

	if err != nil {
		return types.ErrInvalidAsset(localMinDeposit.String())
	}

	// check whether the proposer can afford the initial deposit
	err = common.HasSufficientCoins(proposer, k.bankKeeper.GetCoins(ctx, proposer), initialDeposit)
	if err != nil {
		return types.ErrBalanceNotEnough(proposer.String(), initialDeposit.String())
	}
	return nil
}

// CheckMsgSubmitProposal validates MsgSubmitProposal
func (k Keeper) CheckMsgSubmitProposal(ctx sdk.Context, msg govTypes.MsgSubmitProposal) (sdkErr sdk.Error) {
	switch content := msg.Content.(type) {
	case types.DelistProposal:
		sdkErr = k.checkMsgDelistProposal(ctx, content, msg.Proposer, msg.InitialDeposit)
	default:
		errContent := fmt.Sprintf("unrecognized dex proposal content type: %T", content)
		sdkErr = sdk.ErrUnknownRequest(errContent)
	}
	return
}

// nolint
func (k Keeper) AfterSubmitProposalHandler(ctx sdk.Context, proposal govTypes.Proposal) {}

// VoteHandler handles  delist proposal when voted
func (k Keeper) VoteHandler(ctx sdk.Context, proposal govTypes.Proposal, vote govTypes.Vote) (string, sdk.Error) {
	if _, ok := proposal.Content.(types.DelistProposal); ok {
		delistProposal := proposal.Content.(types.DelistProposal)
		tokenPairName := delistProposal.BaseAsset + "_" + delistProposal.QuoteAsset
		if k.IsTokenPairLocked(ctx, tokenPairName) {
			errContent := fmt.Sprintf("the trading pair (%s) is locked, please retry later", tokenPairName)
			return "", sdk.ErrInternal(errContent)
		}
	}
	return "", nil
}

// RejectedHandler handles delist proposal when rejected
func (k Keeper) RejectedHandler(ctx sdk.Context, content govTypes.Content) {
	if content, ok := content.(types.DelistProposal); ok {
		tokenPairName := fmt.Sprintf("%s_%s", content.BaseAsset, content.QuoteAsset)
		//update the token info from the store
		tokenPair := k.GetTokenPair(ctx, tokenPairName)
		if tokenPair == nil {
			ctx.Logger().Error(fmt.Sprintf("token pair %s does not exist", tokenPairName))
			return
		}
		tokenPair.Delisting = false
		k.UpdateTokenPair(ctx, tokenPairName, tokenPair)
	}
}

// AfterDepositPeriodPassed handles delist proposal when passed
func (k Keeper) AfterDepositPeriodPassed(ctx sdk.Context, proposal govTypes.Proposal) {
	if content, ok := proposal.Content.(types.DelistProposal); ok {
		tokenPairName := fmt.Sprintf("%s_%s", content.BaseAsset, content.QuoteAsset)
		// change the status of the token pair in the store
		tokenPair := k.GetTokenPair(ctx, tokenPairName)
		if tokenPair == nil {
			ctx.Logger().Error(fmt.Sprintf("token pair %s does not exist", tokenPairName))
			return
		}
		tokenPair.Delisting = true
		k.UpdateTokenPair(ctx, tokenPairName, tokenPair)
	}
}

// RemoveFromActiveProposalQueue removes active proposal in queue
func (k Keeper) RemoveFromActiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time) {
	k.govKeeper.RemoveFromActiveProposalQueue(ctx, proposalID, endTime)
}
