package types_test

import (
	"testing"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/store/iavl"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/store/rootmulti"
	storetypes "github.com/FiboChain/fbc/libs/cosmos-sdk/store/types"
	"github.com/FiboChain/fbc/libs/tendermint/types"
	dbm "github.com/FiboChain/fbc/libs/tm-db"

	"github.com/stretchr/testify/suite"
	// "github.com/cosmos/cosmos-sdk/store/iavl"
	// "github.com/cosmos/cosmos-sdk/store/rootmulti"
	// storetypes "github.com/cosmos/cosmos-sdk/store/types"
	// dbm "github.com/tendermint/tm-db"
)

type MerkleTestSuite struct {
	suite.Suite

	store     *rootmulti.Store
	storeKey  *storetypes.KVStoreKey
	iavlStore *iavl.Store
}

func (suite *MerkleTestSuite) SetupTest() {
	types.UnittestOnlySetMilestoneVenus1Height(-1)
	db := dbm.NewMemDB()
	suite.store = rootmulti.NewStore(db)
	suite.storeKey = storetypes.NewKVStoreKey("iavlStoreKey")

	suite.store.MountStoreWithDB(suite.storeKey, storetypes.StoreTypeIAVL, nil)
	suite.store.LoadVersion(0)

	suite.iavlStore = suite.store.GetCommitStore(suite.storeKey).(*iavl.Store)
}

func TestMerkleTestSuite(t *testing.T) {
	suite.Run(t, new(MerkleTestSuite))
}
