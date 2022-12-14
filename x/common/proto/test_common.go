package proto

import (
	"os"
	"testing"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/store"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	abci "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	"github.com/FiboChain/fbc/libs/tendermint/libs/log"
	"github.com/stretchr/testify/require"
	dbm "github.com/FiboChain/fbc/libs/tm-db"
)

func createTestInput(t *testing.T) (sdk.Context, ProtocolKeeper) {
	keyMain := sdk.NewKVStoreKey("main")

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyMain, sdk.StoreTypeIAVL, db)

	require.NoError(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewTMLogger(os.Stdout))

	keeper := NewProtocolKeeper(keyMain)

	return ctx, keeper
}
