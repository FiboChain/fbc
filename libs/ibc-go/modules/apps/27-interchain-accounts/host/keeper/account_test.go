package keeper_test

import (
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	icatypes "github.com/FiboChain/fbc/libs/ibc-go/modules/apps/27-interchain-accounts/types"
	ibctesting "github.com/FiboChain/fbc/libs/ibc-go/testing"
)

func (suite *KeeperTestSuite) TestRegisterInterchainAccount() {
	suite.SetupTest()

	path := NewICAPath(suite.chainA, suite.chainB)
	suite.coordinator.SetupConnections(path)

	// RegisterInterchainAccount
	err := SetupICAPath(path, TestOwnerAddress)
	suite.Require().NoError(err)

	portID, err := icatypes.NewControllerPortID(TestOwnerAddress)
	suite.Require().NoError(err)

	// Get the address of the interchain account stored in state during handshake step
	storedAddr, found := suite.chainB.GetSimApp().ICAHostKeeper.GetInterchainAccountAddress(suite.chainB.GetContext(), ibctesting.FirstConnectionID, portID)
	suite.Require().True(found)

	icaAddr, err := sdk.AccAddressFromBech32(storedAddr)
	suite.Require().NoError(err)

	// Check if account is created
	interchainAccount := suite.chainB.GetSimApp().AccountKeeper.GetAccount(suite.chainB.GetContext(), icaAddr)
	suite.Require().Equal(interchainAccount.GetAddress().String(), storedAddr)
}
