package keeper

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/store"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth"
	"github.com/FiboChain/fbc/libs/tendermint/libs/log"
	"github.com/FiboChain/fbc/x/evm/types"
	"github.com/FiboChain/fbc/x/evm/watcher"
	"github.com/FiboChain/fbc/x/params"
)

// Keeper wraps the CommitStateDB, allowing us to pass in SDK context while adhering
// to the StateDB interface.
type Keeper struct {
	// Amino codec
	cdc *codec.Codec
	// Store key required for the EVM Prefix KVStore. It is required by:
	// - storing Account's Storage State
	// - storing Account's Code
	// - storing transaction Logs
	// - storing block height -> bloom filter map. Needed for the Web3 API.
	// - storing block hash -> block height map. Needed for the Web3 API.
	storeKey sdk.StoreKey
	// Account Keeper for fetching accounts
	accountKeeper types.AccountKeeper
	paramSpace    types.Subspace
	supplyKeeper  types.SupplyKeeper
	bankKeeper    types.BankKeeper
	govKeeper     GovKeeper

	// Transaction counter in a block. Used on StateSB's Prepare function.
	// It is reset to 0 every block on BeginBlock so there's no point in storing the counter
	// on the KVStore or adding it as a field on the EVM genesis state.
	TxCount int
	Bloom   *big.Int
	Bhash   ethcmn.Hash
	LogSize uint
	Watcher *watcher.Watcher
	Ada     types.DbAdapter

	LogsManages *LogsManager

	// add inner block data
	innerBlockData BlockInnerData

	// cache chain config
	cci chainConfigInfo
}

type chainConfigInfo struct {
	// chainConfig cached chain config
	// nil means invalid the cache, we should cache it again.
	cc *types.ChainConfig

	// gasReduced: cached chain config reduces gas costs.
	// when use cached chain config, we restore the gas cost(gasReduced)
	gasReduced sdk.Gas
}

// NewKeeper generates new evm module keeper
func NewKeeper(
	cdc *codec.Codec, storeKey sdk.StoreKey, paramSpace params.Subspace, ak types.AccountKeeper, sk types.SupplyKeeper, bk types.BankKeeper,
	logger log.Logger) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	types.InitTxTraces()
	err := initInnerDB()
	if err != nil {
		panic(err)
	}

	if enable := types.GetEnableBloomFilter(); enable {
		db := types.BloomDb()
		types.InitIndexer(db)
	}
	// NOTE: we pass in the parameter space to the CommitStateDB in order to use custom denominations for the EVM operations
	k := &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: ak,
		paramSpace:    paramSpace,
		supplyKeeper:  sk,
		bankKeeper:    bk,
		TxCount:       0,
		Bloom:         big.NewInt(0),
		LogSize:       0,
		Watcher:       watcher.NewWatcher(logger),
		Ada:           types.DefaultPrefixDb{},

		innerBlockData: defaultBlockInnerData(),
	}
	k.Watcher.SetWatchDataFunc()
	if k.Watcher.Enabled() {
		ak.SetObserverKeeper(k)
	}

	return k
}

// NewKeeper generates new evm module keeper
func NewSimulateKeeper(
	cdc *codec.Codec, storeKey sdk.StoreKey, paramSpace types.Subspace, ak types.AccountKeeper, sk types.SupplyKeeper, bk types.BankKeeper, ada types.DbAdapter,
	logger log.Logger) *Keeper {
	// NOTE: we pass in the parameter space to the CommitStateDB in order to use custom denominations for the EVM operations
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: ak,
		paramSpace:    paramSpace,
		supplyKeeper:  sk,
		bankKeeper:    bk,
		TxCount:       0,
		Bloom:         big.NewInt(0),
		LogSize:       0,
		Watcher:       watcher.NewWatcher(nil),
		Ada:           ada,
	}
}

func (k Keeper) OnAccountUpdated(acc auth.Account) {
	account := acc.GetAddress()
	k.Watcher.DeleteAccount(account)
}

// Logger returns a module-specific logger.
func (k *Keeper) GenerateCSDBParams() types.CommitStateDBParams {
	return types.CommitStateDBParams{
		StoreKey:      k.storeKey,
		ParamSpace:    k.paramSpace,
		AccountKeeper: k.accountKeeper,
		SupplyKeeper:  k.supplyKeeper,
		BankKeeper:    k.bankKeeper,
		Watcher:       k.Watcher,
		Ada:           k.Ada,
		Cdc:           k.cdc,
	}
}

// GeneratePureCSDBParams generates an instance of csdb params ONLY for store setter and getter
func (k Keeper) GeneratePureCSDBParams() types.CommitStateDBParams {
	return types.CommitStateDBParams{
		StoreKey: k.storeKey,
		Watcher:  k.Watcher,
		Ada:      k.Ada,
		Cdc:      k.cdc,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ----------------------------------------------------------------------------
// Block hash mapping functions
// Required by Web3 API.
//  TODO: remove once tendermint support block queries by hash.
// ----------------------------------------------------------------------------

// GetBlockHash gets block height from block consensus hash
func (k Keeper) GetBlockHash(ctx sdk.Context, hash []byte) (int64, bool) {
	store := k.Ada.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixBlockHash)
	bz := store.Get(hash)
	if len(bz) == 0 {
		return 0, false
	}

	height := binary.BigEndian.Uint64(bz)
	return int64(height), true
}

// SetBlockHash sets the mapping from block consensus hash to block height
func (k Keeper) SetBlockHash(ctx sdk.Context, hash []byte, height int64) {
	store := k.Ada.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixBlockHash)
	bz := sdk.Uint64ToBigEndian(uint64(height))
	store.Set(hash, bz)
}

// ----------------------------------------------------------------------------
// Epoch Height -> hash mapping functions
// Required by EVM context's GetHashFunc
// ----------------------------------------------------------------------------

// GetHeightHash returns the block header hash associated with a given block height and chain epoch number.
func (k Keeper) GetHeightHash(ctx sdk.Context, height uint64) common.Hash {
	return types.CreateEmptyCommitStateDB(k.GenerateCSDBParams(), ctx).GetHeightHash(height)
}

// SetHeightHash sets the block header hash associated with a given height.
func (k Keeper) SetHeightHash(ctx sdk.Context, height uint64, hash common.Hash) {
	types.CreateEmptyCommitStateDB(k.GenerateCSDBParams(), ctx).SetHeightHash(height, hash)
}

// ----------------------------------------------------------------------------
// Block bloom bits mapping functions
// Required by Web3 API.
// ----------------------------------------------------------------------------

// GetBlockBloom gets bloombits from block height
func (k Keeper) GetBlockBloom(ctx sdk.Context, height int64) ethtypes.Bloom {
	store := k.Ada.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixBloom)
	has := store.Has(types.BloomKey(height))
	if !has {
		return ethtypes.Bloom{}
	}

	bz := store.Get(types.BloomKey(height))
	return ethtypes.BytesToBloom(bz)
}

func (k Keeper) GetStoreKey() store.StoreKey {
	return k.storeKey
}

// SetBlockBloom sets the mapping from block height to bloom bits
func (k Keeper) SetBlockBloom(ctx sdk.Context, height int64, bloom ethtypes.Bloom) {
	store := k.Ada.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixBloom)
	store.Set(types.BloomKey(height), bloom.Bytes())
}

// GetAccountStorage return state storage associated with an account
func (k Keeper) GetAccountStorage(ctx sdk.Context, address common.Address) (types.Storage, error) {
	storage := types.Storage{}
	csdb := types.CreateEmptyCommitStateDB(k.GenerateCSDBParams(), ctx)
	err := csdb.ForEachStorage(address, func(key, value common.Hash) bool {
		storage = append(storage, types.NewState(key, value))
		return false
	})
	if err != nil {
		return types.Storage{}, err
	}

	return storage, nil
}

// getChainConfig get raw chain config and unmarshal it
func (k *Keeper) getChainConfig(ctx sdk.Context) (types.ChainConfig, bool) {
	// if keeper has cached the chain config, return immediately
	store := k.Ada.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixChainConfig)
	// get from an empty key that's already prefixed by KeyPrefixChainConfig
	bz := store.Get([]byte{})
	if len(bz) == 0 {
		return types.ChainConfig{}, false
	}

	var config types.ChainConfig
	// first 4 bytes are type prefix
	// bz len must > 4; otherwise, MustUnmarshalBinaryBare will panic
	if err := config.UnmarshalFromAmino(k.cdc, bz[4:]); err != nil {
		k.cdc.MustUnmarshalBinaryBare(bz, &config)
	}

	return config, true
}

// GetChainConfig gets chain config, the result if from cached result, or
// it gains chain config and gas costs from getChainConfig, then
// cache the chain config and gas costs.
func (k *Keeper) GetChainConfig(ctx sdk.Context) (types.ChainConfig, bool) {
	// if keeper has cached the chain config, return immediately, and increase gas costs.
	if k.cci.cc != nil {
		ctx.GasMeter().ConsumeGas(k.cci.gasReduced, "cached chain config recover")
		return *k.cci.cc, true
	}

	gasStart := ctx.GasMeter().GasConsumed()
	chainConfig, found := k.getChainConfig(ctx)
	gasStop := ctx.GasMeter().GasConsumed()

	// only cache chain config result when we found it, or try to found again.
	if found {
		k.cci.cc = &chainConfig
		k.cci.gasReduced = gasStop - gasStart
	}

	return chainConfig, found
}

// SetChainConfig sets the mapping from block consensus hash to block height
func (k *Keeper) SetChainConfig(ctx sdk.Context, config types.ChainConfig) {
	store := k.Ada.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixChainConfig)
	bz := k.cdc.MustMarshalBinaryBare(config)
	// get to an empty key that's already prefixed by KeyPrefixChainConfig
	store.Set([]byte{}, bz)

	// invalid the chainConfig
	k.cci.cc = nil
}

// SetGovKeeper sets keeper of gov
func (k *Keeper) SetGovKeeper(gk GovKeeper) {
	k.govKeeper = gk
}

// checks whether the address is blocked
func (k *Keeper) IsAddressBlocked(ctx sdk.Context, addr sdk.AccAddress) bool {
	csdb := types.CreateEmptyCommitStateDB(k.GenerateCSDBParams(), ctx)
	return csdb.GetParams().EnableContractBlockedList && csdb.IsContractInBlockedList(addr.Bytes())
}

func (k *Keeper) IsContractInBlockedList(ctx sdk.Context, addr sdk.AccAddress) bool {
	csdb := types.CreateEmptyCommitStateDB(k.GenerateCSDBParams(), ctx)
	return csdb.IsContractInBlockedList(addr.Bytes())
}
