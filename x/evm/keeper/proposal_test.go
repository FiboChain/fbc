package keeper_test

import (
	"time"

	ethcmn "github.com/ethereum/go-ethereum/common"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/x/evm/types"
	govtypes "github.com/FiboChain/fbc/x/gov/types"
	"github.com/stretchr/testify/require"
)

func (suite *KeeperTestSuite) TestProposal_ManageContractDeploymentWhitelistProposal() {
	addr1 := ethcmn.BytesToAddress([]byte{0x0}).Bytes()
	addr2 := ethcmn.BytesToAddress([]byte{0x1}).Bytes()

	proposal := types.NewManageContractDeploymentWhitelistProposal(
		"default title",
		"default description",
		types.AddressList{addr1, addr2},
		true,
	)

	minDeposit := suite.app.EvmKeeper.GetMinDeposit(suite.ctx, proposal)
	require.Equal(suite.T(), sdk.SysCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(100))}, minDeposit)

	maxDepositPeriod := suite.app.EvmKeeper.GetMaxDepositPeriod(suite.ctx, proposal)
	require.Equal(suite.T(), time.Hour*24, maxDepositPeriod)

	votingPeriod := suite.app.EvmKeeper.GetVotingPeriod(suite.ctx, proposal)
	require.Equal(suite.T(), time.Hour*72, votingPeriod)

	testCases := []struct {
		msg     string
		prepare func()
	}{
		{
			"pass check",
			func() {},
		},
		{
			"pass check when trying to add addresses already exists in whitelist",
			func() {
				suite.stateDB.SetContractDeploymentWhitelist(types.AddressList{addr1, addr2})
			},
		},
		{
			"pass check when trying to delete addresses from whitelist",
			func() {
				proposal.IsAdded = false
			},
		},
		{
			"pass check when trying to delete addresses from whitelist which contains none of them",
			func() {
				// clear whitelist in the store
				suite.stateDB.DeleteContractDeploymentWhitelist(suite.stateDB.GetContractDeploymentWhitelist())
				suite.Require().Zero(len(suite.stateDB.GetContractDeploymentWhitelist()))
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.msg, func() {
			tc.prepare()

			msg := govtypes.NewMsgSubmitProposal(proposal, minDeposit, addr1)
			err := suite.app.EvmKeeper.CheckMsgSubmitProposal(suite.ctx, msg)
			suite.Require().NoError(err)
		})
	}
}

func (suite *KeeperTestSuite) TestProposal_ManageContractBlockedListProposal() {
	addr1 := ethcmn.BytesToAddress([]byte{0x0}).Bytes()
	addr2 := ethcmn.BytesToAddress([]byte{0x1}).Bytes()

	proposal := types.NewManageContractBlockedListProposal(
		"default title",
		"default description",
		types.AddressList{addr1, addr2},
		true,
	)

	minDeposit := suite.app.EvmKeeper.GetMinDeposit(suite.ctx, proposal)
	require.Equal(suite.T(), sdk.SysCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(100))}, minDeposit)

	maxDepositPeriod := suite.app.EvmKeeper.GetMaxDepositPeriod(suite.ctx, proposal)
	require.Equal(suite.T(), time.Hour*24, maxDepositPeriod)

	votingPeriod := suite.app.EvmKeeper.GetVotingPeriod(suite.ctx, proposal)
	require.Equal(suite.T(), time.Hour*72, votingPeriod)

	testCases := []struct {
		msg     string
		prepare func()
	}{
		{
			"pass check",
			func() {},
		},
		{
			"pass check when trying to add addresses already exists in blocked list",
			func() {
				suite.stateDB.SetContractDeploymentWhitelist(types.AddressList{addr1, addr2})
			},
		},
		{
			"pass check when trying to delete addresses from blocked list",
			func() {
				proposal.IsAdded = false
			},
		},
		{
			"pass check when trying to delete addresses from blocked list which contains none of them",
			func() {
				// clear blocked list in the store
				suite.stateDB.DeleteContractBlockedList(suite.stateDB.GetContractBlockedList())
				suite.Require().Zero(len(suite.stateDB.GetContractBlockedList()))
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.msg, func() {
			tc.prepare()

			msg := govtypes.NewMsgSubmitProposal(proposal, minDeposit, addr1)
			err := suite.app.EvmKeeper.CheckMsgSubmitProposal(suite.ctx, msg)
			suite.Require().NoError(err)
		})
	}
}

func (suite *KeeperTestSuite) TestProposal_ManageContractMethodBlockedListProposal() {
	addr1 := ethcmn.BytesToAddress([]byte{0x0}).Bytes()
	addr2 := ethcmn.BytesToAddress([]byte{0x1}).Bytes()
	bcMethodOne1 := types.BlockedContract{
		Address: addr1,
		BlockMethods: types.ContractMethods{
			types.ContractMethod{
				Sign:  "aaaa",
				Extra: "aaaa()",
			},
		},
	}
	bcMethodTwo1 := types.BlockedContract{
		Address: addr2,
		BlockMethods: types.ContractMethods{
			types.ContractMethod{
				Sign:  "aaaa",
				Extra: "aaaa()",
			},
		},
	}
	bcl := types.BlockedContractList{bcMethodOne1, bcMethodTwo1}
	proposal := types.NewManageContractMethodBlockedListProposal(
		"default title",
		"default description",
		bcl,
		true,
	)

	minDeposit := suite.app.EvmKeeper.GetMinDeposit(suite.ctx, proposal)
	require.Equal(suite.T(), sdk.SysCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(100))}, minDeposit)

	maxDepositPeriod := suite.app.EvmKeeper.GetMaxDepositPeriod(suite.ctx, proposal)
	require.Equal(suite.T(), time.Hour*24, maxDepositPeriod)

	votingPeriod := suite.app.EvmKeeper.GetVotingPeriod(suite.ctx, proposal)
	require.Equal(suite.T(), time.Hour*72, votingPeriod)

	testCases := []struct {
		msg     string
		prepare func()
		success bool
	}{
		{
			"pass check",
			func() {},
			true,
		},
		{
			"pass check when trying to add addresses already exists in blocked list",
			func() {
				suite.stateDB.InsertContractMethodBlockedList(bcl)
			},
			true,
		},
		{
			"pass check when trying to delete addresses from blocked list",
			func() {
				proposal.IsAdded = false
			},
			true,
		},
		{
			"pass check when trying to delete addresses from blocked list which is empty",
			func() {
				// clear blocked list in the store
				suite.stateDB.DeleteContractMethodBlockedList(suite.stateDB.GetContractMethodBlockedList())
				suite.Require().Zero(len(suite.stateDB.GetContractBlockedList()))
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.msg, func() {
			tc.prepare()

			msg := govtypes.NewMsgSubmitProposal(proposal, minDeposit, addr1)
			err := suite.app.EvmKeeper.CheckMsgSubmitProposal(suite.ctx, msg)
			if tc.success {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
