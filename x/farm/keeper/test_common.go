package keeper

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/store"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/bank"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/supply"
	abci "github.com/FiboChain/fbc/libs/tendermint/abci/types"
	"github.com/FiboChain/fbc/libs/tendermint/libs/log"
	tmtypes "github.com/FiboChain/fbc/libs/tendermint/types"
	dbm "github.com/FiboChain/fbc/libs/tm-db"
	swap "github.com/FiboChain/fbc/x/ammswap"
	swaptypes "github.com/FiboChain/fbc/x/ammswap/types"
	evm "github.com/FiboChain/fbc/x/evm/keeper"
	evmtypes "github.com/FiboChain/fbc/x/evm/types"
	"github.com/FiboChain/fbc/x/farm/types"
	"github.com/FiboChain/fbc/x/gov"
	govkeeper "github.com/FiboChain/fbc/x/gov/keeper"
	govtypes "github.com/FiboChain/fbc/x/gov/types"
	"github.com/FiboChain/fbc/x/params"
	"github.com/FiboChain/fbc/x/token"
	"github.com/stretchr/testify/require"
)

const (
	TestChainID = "fbchain"
)

var (
	Addrs = createTestAddrs(10)
)

type MockFarmKeeper struct {
	Keeper
	StoreKey       sdk.StoreKey
	SupplyKeeper   supply.Keeper
	MountedStore   store.MultiStore
	AccKeeper      auth.AccountKeeper
	BankKeeper     bank.Keeper
	TokenKeeper    token.Keeper
	SwapKeeper     swap.Keeper
	ObserverKeeper *MockObserverKeeper
}

type MockObserverKeeper struct {
	ObserverData MockObserverData
}

func (ok *MockObserverKeeper) OnFarmClaim(ctx sdk.Context, address sdk.AccAddress, poolName string, claimedCoins sdk.SysCoins) {
	ok.ObserverData = MockObserverData{
		Address:      address,
		PoolName:     poolName,
		ClaimedCoins: claimedCoins,
	}
}

type MockObserverData struct {
	Address      sdk.AccAddress
	PoolName     string
	ClaimedCoins sdk.SysCoins
}

func NewMockFarmKeeper(
	k Keeper, keyStoreKey sdk.StoreKey, sKeeper supply.Keeper,
	ms store.MultiStore, accKeeper auth.AccountKeeper, bankKeeper bank.Keeper,
	tokenKeeper token.Keeper, swapKeeper swap.Keeper, observerKeeper *MockObserverKeeper,
) MockFarmKeeper {
	return MockFarmKeeper{
		k,
		keyStoreKey,
		sKeeper,
		ms,
		accKeeper,
		bankKeeper,
		tokenKeeper,
		swapKeeper,
		observerKeeper,
	}
}

func GetKeeper(t *testing.T) (sdk.Context, MockFarmKeeper) {
	// 0.1 init store key
	keyFarm := sdk.NewKVStoreKey(types.StoreKey)
	tkeyFarm := sdk.NewTransientStoreKey(types.TStoreKey)
	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)
	keyToken := sdk.NewKVStoreKey(token.StoreKey)
	keyLock := sdk.NewKVStoreKey(token.KeyLock)
	keySwap := sdk.NewKVStoreKey(swaptypes.StoreKey)
	keyEvm := sdk.NewKVStoreKey(evmtypes.StoreKey)
	keyGov := sdk.NewKVStoreKey(govtypes.StoreKey)

	// 0.2 init db
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(tkeyFarm, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyFarm, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyToken, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySwap, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyGov, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	// 0.3 init context
	ctx := sdk.NewContext(ms, abci.Header{ChainID: TestChainID}, false, log.NewNopLogger())
	ctx = ctx.WithConsensusParams(
		&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519},
			},
		},
	)

	// 0.4 init codec
	cdc := codec.New()
	types.RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	token.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	params.RegisterCodec(cdc)
	swaptypes.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	// 1.1 init param keeper
	pk := params.NewKeeper(cdc, keyParams, tkeyParams)

	// 1.2 init account keeper
	ak := auth.NewAccountKeeper(cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)

	// 1.3 init bank keeper
	feeCollectorAcc := supply.NewEmptyModuleAccount(auth.FeeCollectorName)
	farmAcc := supply.NewEmptyModuleAccount(types.ModuleName, supply.Burner, supply.Minter)
	yieldFarmingAccount := supply.NewEmptyModuleAccount(types.YieldFarmingAccount, supply.Burner, supply.Minter)
	mintFarmingAccount := supply.NewEmptyModuleAccount(types.MintFarmingAccount, supply.Burner, supply.Minter)
	swapModuleAccount := supply.NewEmptyModuleAccount(swap.ModuleName, supply.Burner, supply.Minter)

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[feeCollectorAcc.String()] = true
	blacklistedAddrs[farmAcc.String()] = true

	bk := bank.NewBaseKeeper(ak, pk.Subspace(bank.DefaultParamspace), blacklistedAddrs)
	// fill all the addresses with some coins
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000)))
	for _, addr := range Addrs {
		_, err := bk.AddCoins(ctx, addr, initCoins)
		if err != nil {
			panic(err)
		}
	}

	// 1.4 init supply keeper
	maccPerms := map[string][]string{
		auth.FeeCollectorName:     nil,
		types.ModuleName:          nil,
		types.YieldFarmingAccount: nil,
		types.MintFarmingAccount:  nil,
		swap.ModuleName:           {supply.Burner, supply.Minter},
		govtypes.ModuleName:       nil,
	}
	sk := supply.NewKeeper(cdc, keySupply, ak, bk, maccPerms)
	sk.SetSupply(ctx, supply.NewSupply(sdk.NewDecCoinsFromDec(sdk.DefaultBondDenom, sdk.NewDec(1000000000))))
	sk.SetModuleAccount(ctx, feeCollectorAcc)
	sk.SetModuleAccount(ctx, farmAcc)
	sk.SetModuleAccount(ctx, yieldFarmingAccount)
	sk.SetModuleAccount(ctx, mintFarmingAccount)
	sk.SetModuleAccount(ctx, swapModuleAccount)

	// 1.5 init token keeper
	tk := token.NewKeeper(bk, pk.Subspace(token.DefaultParamspace), auth.FeeCollectorName, sk, keyToken, keyLock, cdc, false, ak)

	// 1.6 init swap keeper
	swapKeeper := swap.NewKeeper(sk, tk, cdc, keySwap, pk.Subspace(swaptypes.DefaultParamspace))
	evmKeeper := evm.NewKeeper(cdc, keyEvm, pk.Subspace(evmtypes.DefaultParamspace), &ak, sk, bk, log.NewNopLogger())

	// 1.7 init farm keeper
	fk := NewKeeper(auth.FeeCollectorName, sk, tk, swapKeeper, *evmKeeper, pk.Subspace(types.DefaultParamspace), keyFarm, cdc)

	fk.SetParams(ctx, types.DefaultParams())

	// 1.8 init gov keeper
	govSubspace := pk.Subspace(govtypes.DefaultParamspace)
	govProposalHandlerRouter := govkeeper.NewProposalHandlerRouter()
	govProposalHandlerRouter.AddRoute(params.RouterKey, pk)
	govKeeper := gov.NewKeeper(cdc, keyGov, pk, govSubspace, sk, nil,
		types.DefaultCodespace, gov.NewRouter(), bk, govProposalHandlerRouter, auth.FeeCollectorName)
	minDeposit := sdk.NewDecCoinsFromDec(sdk.DefaultBondDenom, sdk.NewDec(100))
	depositParams := govtypes.DepositParams{
		MinDeposit:       minDeposit,
		MaxDepositPeriod: time.Hour * 24,
	}
	votingParams := govtypes.VotingParams{
		VotingPeriod: time.Hour * 72,
	}
	tallyParams := govtypes.TallyParams{
		Quorum:          sdk.NewDecWithPrec(334, 3),
		Threshold:       sdk.NewDecWithPrec(5, 1),
		Veto:            sdk.NewDecWithPrec(334, 3),
		YesInVotePeriod: sdk.NewDecWithPrec(667, 3),
	}
	govKeeper.SetDepositParams(ctx, depositParams)
	govKeeper.SetVotingParams(ctx, votingParams)
	govKeeper.SetTallyParams(ctx, tallyParams)

	fk.SetGovKeeper(govKeeper)

	observerKeeper := &MockObserverKeeper{}
	fk.SetObserverKeeper(observerKeeper)
	// 2. init mock keeper
	mk := NewMockFarmKeeper(fk, keyFarm, sk, ms, ak, bk, tk, swapKeeper, observerKeeper)

	//// 3. init mockApp
	//mApp := mock.NewApp()
	//mApp.Router().AddRoute(types.RouterKey, NewHandler(fk))
	//
	//require.NoError(t, mApp.CompleteSetup(mk.StoreKey))
	return ctx, mk
}

func createTestAddrs(numAddrs int) []sdk.AccAddress {
	var addresses []sdk.AccAddress
	var buffer bytes.Buffer

	// start at 100 so we can make up to 999 test addresses with valid test addresses
	for i := 100; i < (numAddrs + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6") //base address string

		buffer.WriteString(numString) //adding on final two digits to make addresses unique
		res, err := sdk.AccAddressFromHex(buffer.String())
		if err != nil {
			fmt.Print("error")
		}
		bech := res.String()
		addresses = append(addresses, testAddr(buffer.String(), bech))
		buffer.Reset()
	}
	return addresses
}

// TestAddr is designed for incode address generation
func testAddr(addr string, bech string) sdk.AccAddress {

	res, err := sdk.AccAddressFromHex(addr)
	if err != nil {
		panic(err)
	}
	bechexpected := res.String()
	if bech != bechexpected {
		panic("Bech encoding doesn't match reference")
	}

	bechres, err := sdk.AccAddressFromBech32(bech)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(bechres, res) {
		panic("Bech decode and hex decode don't match")
	}

	return res
}

func initPoolsAndLockInfos(
	t *testing.T, ctx sdk.Context, mockKeeper MockFarmKeeper,
) (pools types.FarmPools, lockInfos []types.LockInfo) {
	pool1Name := "pool1"
	pool2Name := "pool2"

	pool1LockedAmount := sdk.NewDecCoin("xxb", sdk.NewInt(100))
	pool2LockedAmount := sdk.NewDecCoin("yyb", sdk.NewInt(100))

	lockInfos = []types.LockInfo{
		types.NewLockInfo(Addrs[0], pool1Name, pool1LockedAmount, 80, 1),
		types.NewLockInfo(Addrs[1], pool1Name, pool1LockedAmount, 90, 2),
		types.NewLockInfo(Addrs[0], pool2Name, pool2LockedAmount, 80, 1),
		types.NewLockInfo(Addrs[1], pool2Name, pool2LockedAmount, 90, 2),
	}

	for _, lockInfo := range lockInfos {
		mockKeeper.Keeper.SetLockInfo(ctx, lockInfo)
		mockKeeper.Keeper.SetAddressInFarmPool(ctx, lockInfo.PoolName, lockInfo.Owner)
	}

	yieldAmount := sdk.NewDecCoin("wwb", sdk.NewInt(1000))
	poolYieldedInfos := types.YieldedTokenInfos{
		types.NewYieldedTokenInfo(yieldAmount, 100, sdk.NewDec(10)),
	}
	pools = types.FarmPools{
		types.NewFarmPool(
			Addrs[2], pool1Name, sdk.NewDecCoinFromDec(pool1LockedAmount.Denom, sdk.ZeroDec()),
			sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
			pool1LockedAmount.Add(pool1LockedAmount), poolYieldedInfos, sdk.SysCoins(nil),
		),
		types.NewFarmPool(
			Addrs[3], pool2Name, sdk.NewDecCoinFromDec(pool2LockedAmount.Denom, sdk.ZeroDec()),
			sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.NewInt(200)),
			pool2LockedAmount.Add(pool2LockedAmount), poolYieldedInfos, sdk.SysCoins(nil),
		),
	}
	for _, pool := range pools {
		mockKeeper.Keeper.SetFarmPool(ctx, pool)
		mockKeeper.Keeper.SetPoolHistoricalRewards(
			ctx, pool.Name, 1, types.NewPoolHistoricalRewards(sdk.SysCoins{}, 1),
		)
		mockKeeper.Keeper.SetPoolHistoricalRewards(
			ctx, pool.Name, 2, types.NewPoolHistoricalRewards(sdk.SysCoins{}, 2),
		)
		mockKeeper.Keeper.SetPoolCurrentRewards(
			ctx, pool.Name, types.NewPoolCurrentRewards(90, 3, sdk.SysCoins{}),
		)

		moduleAcc := mockKeeper.supplyKeeper.GetModuleAccount(ctx, types.ModuleName)
		err := moduleAcc.SetCoins(
			moduleAcc.GetCoins().Add2(sdk.SysCoins{pool.DepositAmount}).Add2(sdk.SysCoins{pool.TotalValueLocked}),
		)
		require.Nil(t, err)
		mockKeeper.supplyKeeper.SetModuleAccount(ctx, moduleAcc)

		yieldAcc := mockKeeper.supplyKeeper.GetModuleAccount(ctx, types.YieldFarmingAccount)
		err = yieldAcc.SetCoins(
			yieldAcc.GetCoins().Add2(sdk.SysCoins{pool.YieldedTokenInfos[0].RemainingAmount}).
				Add2(pool.TotalAccumulatedRewards),
		)
		require.Nil(t, err)
		mockKeeper.supplyKeeper.SetModuleAccount(ctx, yieldAcc)
	}
	mockKeeper.Keeper.SetWhitelist(ctx, pools[0].Name)
	return
}

var _ govtypes.Content = MockContent{}

type MockContent struct{}

func (m MockContent) GetTitle() string {
	return ""
}

func (m MockContent) GetDescription() string {
	return ""
}

func (m MockContent) ProposalRoute() string {
	return ""
}

func (m MockContent) ProposalType() string {
	return ""
}

func (m MockContent) ValidateBasic() sdk.Error {
	return nil
}

func (m MockContent) String() string {
	return ""
}

func SetSwapTokenPair(ctx sdk.Context, k Keeper, token0Symbol, token1Symbol string) {
	pairName := swaptypes.GetSwapTokenPairName(token0Symbol, token1Symbol)
	tokenPair := swaptypes.NewSwapPair(token0Symbol, token1Symbol)
	k.swapKeeper.SetSwapTokenPair(ctx, pairName, tokenPair)
}
