package types

import (
	"fmt"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/codec"
	"math/big"
	"sort"
	"sync"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/x/auth"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/store/prefix"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/store/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethvm "github.com/ethereum/go-ethereum/core/vm"
	ethermint "github.com/FiboChain/fbc/app/types"
	sdk "github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	"github.com/FiboChain/fbc/x/common/analyzer"
	"github.com/FiboChain/fbc/x/params"
)

var (
	_ ethvm.StateDB = (*CommitStateDB)(nil)

	zeroBalance = sdk.ZeroInt().BigInt()
)

type revision struct {
	id           int
	journalIndex int
}

type CommitStateDBParams struct {
	StoreKey      sdk.StoreKey
	ParamSpace    Subspace
	AccountKeeper AccountKeeper
	SupplyKeeper  SupplyKeeper
	Watcher       Watcher
	BankKeeper    BankKeeper
	Ada           DbAdapter
	// Amino codec
	Cdc *codec.Codec
}

type Watcher interface {
	SaveAccount(account auth.Account, isDirectly bool)
	AddDelAccMsg(account auth.Account, isDirectly bool)
	SaveState(addr ethcmn.Address, key, value []byte)
	Enabled() bool
	SaveContractBlockedListItem(addr sdk.AccAddress)
	SaveContractDeploymentWhitelistItem(addr sdk.AccAddress)
	DeleteContractBlockedList(addr sdk.AccAddress)
	DeleteContractDeploymentWhitelist(addr sdk.AccAddress)
	SaveContractMethodBlockedListItem(addr sdk.AccAddress, methods []byte)
}

type CacheCode struct {
	CodeHash []byte
	Code     []byte
}

// CommitStateDB implements the Geth state.StateDB interface. Instead of using
// a trie and database for querying and persistence, the Keeper uses KVStores
// and an account mapper is used to facilitate state transitions.
//
// TODO: This implementation is subject to change in regards to its statefull
// manner. In otherwords, how this relates to the keeper in this module.
type CommitStateDB struct {
	// TODO: We need to store the context as part of the structure itself opposed
	// to being passed as a parameter (as it should be) in order to implement the
	// StateDB interface. Perhaps there is a better way.
	ctx sdk.Context

	storeKey      sdk.StoreKey
	paramSpace    Subspace
	accountKeeper AccountKeeper
	supplyKeeper  SupplyKeeper
	Watcher       Watcher
	bankKeeper    BankKeeper

	// array that hold 'live' objects, which will get modified while processing a
	// state transition
	stateObjects      map[ethcmn.Address]*stateEntry
	stateObjectsDirty map[ethcmn.Address]struct{}

	// The refund counter, also used by state transitioning.
	refund uint64

	thash, bhash ethcmn.Hash
	txIndex      int
	logSize      uint

	logs []*ethtypes.Log

	// TODO: Determine if we actually need this as we do not need preimages in
	// the SDK, but it seems to be used elsewhere in Geth.
	preimages           []preimageEntry
	hashToPreimageIndex map[ethcmn.Hash]int // map from hash to the index of the preimages slice

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memo-ized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        *journal
	validRevisions []revision
	nextRevisionID int

	// Per-transaction access list
	accessList *accessList

	// mutex for state deep copying
	lock sync.Mutex

	params *Params

	codeCache map[ethcmn.Address]CacheCode

	dbAdapter DbAdapter

	// Amino codec
	cdc *codec.Codec
}

type StoreProxy interface {
	Set(key, value []byte)
	Get(key []byte) []byte
	Delete(key []byte)
	Has(key []byte) bool
}

type DbAdapter interface {
	NewStore(parent types.KVStore, prefix []byte) StoreProxy
}

type DefaultPrefixDb struct {
}

func (d DefaultPrefixDb) NewStore(parent types.KVStore, Prefix []byte) StoreProxy {
	return prefix.NewStore(parent, Prefix)
}

// newCommitStateDB returns a reference to a newly initialized CommitStateDB
// which implements Geth's state.StateDB interface.
//
// CONTRACT: Stores used for state must be cache-wrapped as the ordering of the
// key/value space matters in determining the merkle root.
func newCommitStateDB(
	ctx sdk.Context, storeKey sdk.StoreKey, paramSpace params.Subspace, ak AccountKeeper, sk SupplyKeeper, bk BankKeeper, watcher Watcher,
	cdc *codec.Codec) *CommitStateDB {
	return &CommitStateDB{
		ctx:                 ctx,
		storeKey:            storeKey,
		paramSpace:          paramSpace,
		accountKeeper:       ak,
		supplyKeeper:        sk,
		bankKeeper:          bk,
		Watcher:             watcher,
		stateObjects:        make(map[ethcmn.Address]*stateEntry),
		stateObjectsDirty:   make(map[ethcmn.Address]struct{}),
		preimages:           []preimageEntry{},
		hashToPreimageIndex: make(map[ethcmn.Hash]int),
		journal:             newJournal(),
		validRevisions:      []revision{},
		accessList:          newAccessList(),
		logs:                []*ethtypes.Log{},
		codeCache:           make(map[ethcmn.Address]CacheCode, 0),
		dbAdapter:           DefaultPrefixDb{},
		cdc:                 cdc,
	}
}

func CreateEmptyCommitStateDB(csdbParams CommitStateDBParams, ctx sdk.Context) *CommitStateDB {
	return &CommitStateDB{
		ctx: ctx,

		storeKey:      csdbParams.StoreKey,
		paramSpace:    csdbParams.ParamSpace,
		accountKeeper: csdbParams.AccountKeeper,
		supplyKeeper:  csdbParams.SupplyKeeper,
		bankKeeper:    csdbParams.BankKeeper,
		Watcher:       csdbParams.Watcher,
		cdc:           csdbParams.Cdc,

		stateObjects:        make(map[ethcmn.Address]*stateEntry),
		stateObjectsDirty:   make(map[ethcmn.Address]struct{}),
		preimages:           []preimageEntry{},
		hashToPreimageIndex: make(map[ethcmn.Hash]int),
		journal:             newJournal(),
		validRevisions:      []revision{},
		accessList:          newAccessList(),
		logSize:             0,
		logs:                []*ethtypes.Log{},
		codeCache:           make(map[ethcmn.Address]CacheCode, 0),
		dbAdapter:           csdbParams.Ada,
	}
}

func (csdb *CommitStateDB) SetInternalDb(dba DbAdapter) {
	csdb.dbAdapter = dba
}

// WithContext returns a Database with an updated SDK context
func (csdb *CommitStateDB) WithContext(ctx sdk.Context) *CommitStateDB {
	csdb.ctx = ctx
	return csdb
}

func (csdb *CommitStateDB) GetCacheCode(addr ethcmn.Address) *CacheCode {
	if !csdb.ctx.IsCheckTx() {
		funcName := "GetCacheCode"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	code, ok := csdb.codeCache[addr]
	if ok {
		return &code
	}

	return nil
}

func (csdb *CommitStateDB) IteratorCode(cb func(addr ethcmn.Address, c CacheCode) bool) {
	for addr, v := range csdb.codeCache {
		cb(addr, v)
	}
}

// ----------------------------------------------------------------------------
// Setters
// ----------------------------------------------------------------------------

// SetHeightHash sets the block header hash associated with a given height.
func (csdb *CommitStateDB) SetHeightHash(height uint64, hash ethcmn.Hash) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "SetHeightHash"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	store := csdb.dbAdapter.NewStore(csdb.ctx.KVStore(csdb.storeKey), KeyPrefixHeightHash)
	key := HeightHashKey(height)
	store.Set(key, hash.Bytes())
}

// SetParams sets the evm parameters to the param space.
func (csdb *CommitStateDB) SetParams(params Params) {
	csdb.params = &params
	csdb.paramSpace.SetParamSet(csdb.ctx, &params)
	GetEvmParamsCache().SetNeedParamsUpdate()
}

// SetBalance sets the balance of an account.
func (csdb *CommitStateDB) SetBalance(addr ethcmn.Address, amount *big.Int) {
	so := csdb.GetOrNewStateObject(addr)
	if so != nil {
		so.SetBalance(amount)
	}
}

// AddBalance adds amount to the account associated with addr.
func (csdb *CommitStateDB) AddBalance(addr ethcmn.Address, amount *big.Int) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "AddBalance"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.GetOrNewStateObject(addr)
	if so != nil {
		so.AddBalance(amount)
	}
}

// SubBalance subtracts amount from the account associated with addr.
func (csdb *CommitStateDB) SubBalance(addr ethcmn.Address, amount *big.Int) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "SubBalance"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.GetOrNewStateObject(addr)
	if so != nil {
		so.SubBalance(amount)
	}
}

// SetNonce sets the nonce (sequence number) of an account.
func (csdb *CommitStateDB) SetNonce(addr ethcmn.Address, nonce uint64) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "SetNonce"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.GetOrNewStateObject(addr)
	if so != nil {
		so.SetNonce(nonce)
	}
}

// SetState sets the storage state with a key, value pair for an account.
func (csdb *CommitStateDB) SetState(addr ethcmn.Address, key, value ethcmn.Hash) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "SetState"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.GetOrNewStateObject(addr)
	if so != nil {
		so.SetState(nil, key, value)
	}
}

// SetCode sets the code for a given account.
func (csdb *CommitStateDB) SetCode(addr ethcmn.Address, code []byte) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "SetCode"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.GetOrNewStateObject(addr)
	hash := Keccak256HashWithCache(code)
	if so != nil {
		so.SetCode(hash, code)
		csdb.codeCache[addr] = CacheCode{
			CodeHash: hash.Bytes(),
			Code:     code,
		}
	}
}

// ----------------------------------------------------------------------------
// Transaction logs
// Required for upgrade logic or ease of querying.
// NOTE: we use BinaryLengthPrefixed since the tx logs are also included on Result data,
// which can't use BinaryBare.
// ----------------------------------------------------------------------------

// SetLogs sets the logs for a transaction in the KVStore.
func (csdb *CommitStateDB) SetLogs(logs []*ethtypes.Log) {
	csdb.logs = logs
}

// DeleteLogs removes the logs from the KVStore. It is used during journal.Revert.
func (csdb *CommitStateDB) DeleteLogs() {
	csdb.logs = []*ethtypes.Log{}
}

// AddLog adds a new log to the state and sets the log metadata from the state.
func (csdb *CommitStateDB) AddLog(log *ethtypes.Log) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "AddLog"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	csdb.journal.append(addLogChange{txhash: csdb.thash})

	log.TxHash = csdb.thash
	log.BlockHash = csdb.bhash
	log.TxIndex = uint(csdb.txIndex)
	log.Index = csdb.logSize

	csdb.logSize = csdb.logSize + 1
	csdb.logs = append(csdb.logs, log)
}

// AddPreimage records a SHA3 preimage seen by the VM.
func (csdb *CommitStateDB) AddPreimage(hash ethcmn.Hash, preimage []byte) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "AddPreimage"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	if _, ok := csdb.hashToPreimageIndex[hash]; !ok {
		csdb.journal.append(addPreimageChange{hash: hash})

		pi := make([]byte, len(preimage))
		copy(pi, preimage)

		csdb.preimages = append(csdb.preimages, preimageEntry{hash: hash, preimage: pi})
		csdb.hashToPreimageIndex[hash] = len(csdb.preimages) - 1
	}
}

// AddRefund adds gas to the refund counter.
func (csdb *CommitStateDB) AddRefund(gas uint64) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "AddRefund"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	csdb.journal.append(refundChange{prev: csdb.refund})
	csdb.refund += gas
}

// SubRefund removes gas from the refund counter. It will panic if the refund
// counter goes below zero.
func (csdb *CommitStateDB) SubRefund(gas uint64) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "SubRefund"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	csdb.journal.append(refundChange{prev: csdb.refund})
	if gas > csdb.refund {
		panic("refund counter below zero")
	}

	csdb.refund -= gas
}

// AddAddressToAccessList adds the given address to the access list
func (csdb *CommitStateDB) AddAddressToAccessList(addr ethcmn.Address) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "AddAddressToAccessList"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	if csdb.accessList.AddAddress(addr) {
		csdb.journal.append(accessListAddAccountChange{&addr})
	}
}

// AddSlotToAccessList adds the given (address, slot)-tuple to the access list
func (csdb *CommitStateDB) AddSlotToAccessList(addr ethcmn.Address, slot ethcmn.Hash) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "AddSlotToAccessList"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	addrMod, slotMod := csdb.accessList.AddSlot(addr, slot)
	if addrMod {
		// In practice, this should not happen, since there is no way to enter the
		// scope of 'address' without having the 'address' become already added
		// to the access list (via call-variant, create, etc).
		// Better safe than sorry, though
		csdb.journal.append(accessListAddAccountChange{&addr})
	}
	if slotMod {
		csdb.journal.append(accessListAddSlotChange{
			address: &addr,
			slot:    &slot,
		})
	}
}
func (csdb *CommitStateDB) PrepareAccessList(sender ethcmn.Address, dest *ethcmn.Address, precompiles []ethcmn.Address, txAccesses ethtypes.AccessList) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "PrepareAccessList"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	csdb.AddAddressToAccessList(sender)
	if csdb != nil {
		csdb.AddAddressToAccessList(*dest)
		// If it's a create-tx, the destination will be added inside evm.create
	}
	for _, addr := range precompiles {
		csdb.AddAddressToAccessList(addr)
	}
	for _, el := range txAccesses {
		csdb.AddAddressToAccessList(el.Address)
		for _, key := range el.StorageKeys {
			csdb.AddSlotToAccessList(el.Address, key)
		}
	}
}

// AddressInAccessList returns true if the given address is in the access list.
func (csdb *CommitStateDB) AddressInAccessList(addr ethcmn.Address) bool {
	if !csdb.ctx.IsCheckTx() {
		funcName := "AddressInAccessList"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	return csdb.accessList.ContainsAddress(addr)
}

// SlotInAccessList returns true if the given (address, slot)-tuple is in the access list.
func (csdb *CommitStateDB) SlotInAccessList(addr ethcmn.Address, slot ethcmn.Hash) (bool, bool) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "SlotInAccessList"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	return csdb.accessList.Contains(addr, slot)
}

// ----------------------------------------------------------------------------
// Getters
// ----------------------------------------------------------------------------

// GetHeightHash returns the block header hash associated with a given block height and chain epoch number.
func (csdb *CommitStateDB) GetHeightHash(height uint64) ethcmn.Hash {
	store := csdb.dbAdapter.NewStore(csdb.ctx.KVStore(csdb.storeKey), KeyPrefixHeightHash)
	key := HeightHashKey(height)
	bz := store.Get(key)
	if len(bz) == 0 {
		return ethcmn.Hash{}
	}

	return ethcmn.BytesToHash(bz)
}

// GetParams returns the total set of evm parameters.
func (csdb *CommitStateDB) GetParams() Params {
	if csdb.params == nil {
		var params Params
		if csdb.ctx.IsDeliver() {
			if GetEvmParamsCache().IsNeedParamsUpdate() {
				csdb.paramSpace.GetParamSet(csdb.ctx, &params)
				GetEvmParamsCache().UpdateParams(params)
			} else {
				params = GetEvmParamsCache().GetParams()
			}
		} else {
			csdb.paramSpace.GetParamSet(csdb.ctx, &params)
		}
		csdb.params = &params
	}
	return *csdb.params
}

// GetBalance retrieves the balance from the given address or 0 if object not
// found.
func (csdb *CommitStateDB) GetBalance(addr ethcmn.Address) *big.Int {
	if !csdb.ctx.IsCheckTx() {
		funcName := "GetBalance"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.getStateObject(addr)
	if so != nil {
		return so.Balance()
	}

	return zeroBalance
}

// GetNonce returns the nonce (sequence number) for a given account.
func (csdb *CommitStateDB) GetNonce(addr ethcmn.Address) uint64 {
	if !csdb.ctx.IsCheckTx() {
		funcName := "GetNonce"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.getStateObject(addr)
	if so != nil {
		return so.Nonce()
	}

	return 0
}

// TxIndex returns the current transaction index set by Prepare.
func (csdb *CommitStateDB) TxIndex() int {
	return csdb.txIndex
}

// BlockHash returns the current block hash set by Prepare.
func (csdb *CommitStateDB) BlockHash() ethcmn.Hash {
	return csdb.bhash
}

func (csdb *CommitStateDB) SetBlockHash(hash ethcmn.Hash) {
	csdb.bhash = hash
}

// GetCode returns the code for a given account.
func (csdb *CommitStateDB) GetCode(addr ethcmn.Address) []byte {
	if !csdb.ctx.IsCheckTx() {
		funcName := "GetCode"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	// check for the contract calling from blocked list if contract blocked list is enabled
	if csdb.GetParams().EnableContractBlockedList && csdb.IsContractInBlockedList(addr.Bytes()) {
		err := ErrContractBlockedVerify{fmt.Sprintf("failed. the contract %s is not allowed to invoke", addr.Hex())}
		panic(err)
	}

	so := csdb.getStateObject(addr)
	if so != nil {
		return so.Code(nil)
	}

	return nil
}

// GetCode returns the code for a given code hash.
func (csdb *CommitStateDB) GetCodeByHash(hash ethcmn.Hash) []byte {
	ctx := csdb.ctx
	store := csdb.dbAdapter.NewStore(ctx.KVStore(csdb.storeKey), KeyPrefixCode)
	code := store.Get(hash.Bytes())

	return code
}

// GetCodeSize returns the code size for a given account.
func (csdb *CommitStateDB) GetCodeSize(addr ethcmn.Address) int {
	if !csdb.ctx.IsCheckTx() {
		funcName := "GetCodeSize"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.getStateObject(addr)
	if so == nil {
		return 0
	}

	if so.code != nil {
		return len(so.code)
	}

	return len(so.Code(nil))
}

// GetCodeHash returns the code hash for a given account.
func (csdb *CommitStateDB) GetCodeHash(addr ethcmn.Address) ethcmn.Hash {
	if !csdb.ctx.IsCheckTx() {
		funcName := "GetCodeHash"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.getStateObject(addr)
	if so == nil {
		return ethcmn.Hash{}
	}

	return ethcmn.BytesToHash(so.CodeHash())
}

// GetState retrieves a value from the given account's storage store.
func (csdb *CommitStateDB) GetState(addr ethcmn.Address, hash ethcmn.Hash) ethcmn.Hash {
	if !csdb.ctx.IsCheckTx() {
		funcName := "GetState"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.getStateObject(addr)
	if so != nil {
		return so.GetState(nil, hash)
	}

	return ethcmn.Hash{}
}

// GetStateByKey retrieves a value from the given account's storage store.
func (csdb *CommitStateDB) GetStateByKey(addr ethcmn.Address, hash ethcmn.Hash) ethcmn.Hash {
	ctx := csdb.ctx
	store := csdb.dbAdapter.NewStore(ctx.KVStore(csdb.storeKey), AddressStoragePrefix(addr))
	data := store.Get(hash.Bytes())

	return ethcmn.BytesToHash(data)
}

// GetCommittedState retrieves a value from the given account's committed
// storage.
func (csdb *CommitStateDB) GetCommittedState(addr ethcmn.Address, hash ethcmn.Hash) ethcmn.Hash {
	if !csdb.ctx.IsCheckTx() {
		funcName := "GetCommittedState"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.getStateObject(addr)
	if so != nil {
		return so.GetCommittedState(nil, hash)
	}

	return ethcmn.Hash{}
}

// GetLogs returns the current logs for a given transaction hash from the KVStore.
func (csdb *CommitStateDB) GetLogs() []*ethtypes.Log {
	return csdb.logs
}

// GetRefund returns the current value of the refund counter.
func (csdb *CommitStateDB) GetRefund() uint64 {
	if !csdb.ctx.IsCheckTx() {
		funcName := "GetRefund"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	return csdb.refund
}

// Preimages returns a list of SHA3 preimages that have been submitted.
func (csdb *CommitStateDB) Preimages() map[ethcmn.Hash][]byte {
	preimages := map[ethcmn.Hash][]byte{}

	for _, pe := range csdb.preimages {
		preimages[pe.hash] = pe.preimage
	}
	return preimages
}

// HasSuicided returns if the given account for the specified address has been
// killed.
func (csdb *CommitStateDB) HasSuicided(addr ethcmn.Address) bool {
	if !csdb.ctx.IsCheckTx() {
		funcName := "HasSuicided"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.getStateObject(addr)
	if so != nil {
		return so.suicided
	}

	return false
}

// StorageTrie returns nil as the state in Ethermint does not use a direct
// storage trie.
func (csdb *CommitStateDB) StorageTrie(addr ethcmn.Address) ethstate.Trie {
	return nil
}

// ----------------------------------------------------------------------------
// Persistence
// ----------------------------------------------------------------------------

// Commit writes the state to the appropriate KVStores. For each state object
// in the cache, it will either be removed, or have it's code set and/or it's
// state (storage) updated. In addition, the state object (account) itself will
// be written. Finally, the root hash (version) will be returned.
func (csdb *CommitStateDB) Commit(deleteEmptyObjects bool) (ethcmn.Hash, error) {
	defer csdb.clearJournalAndRefund()

	// remove dirty state object entries based on the journal
	for _, dirty := range csdb.journal.dirties {
		csdb.stateObjectsDirty[dirty.address] = struct{}{}
	}

	// set the state objects
	for _, stateEntry := range csdb.stateObjects {
		_, isDirty := csdb.stateObjectsDirty[stateEntry.address]

		switch {
		case stateEntry.stateObject.suicided || (isDirty && deleteEmptyObjects && stateEntry.stateObject.empty()):
			// If the state object has been removed, don't bother syncing it and just
			// remove it from the store.
			csdb.deleteStateObject(stateEntry.stateObject)

		case isDirty:
			// write any contract code associated with the state object
			if stateEntry.stateObject.code != nil && stateEntry.stateObject.dirtyCode {
				stateEntry.stateObject.commitCode()
				stateEntry.stateObject.dirtyCode = false
			}

			// update the object in the KVStore
			if err := csdb.updateStateObject(stateEntry.stateObject); err != nil {
				return ethcmn.Hash{}, err
			}
		}

		delete(csdb.stateObjectsDirty, stateEntry.address)
	}

	// NOTE: Ethereum returns the trie merkle root here, but as commitment
	// actually happens in the BaseApp at EndBlocker, we do not know the root at
	// this time.
	return ethcmn.Hash{}, nil
}

// Finalise finalizes the state objects (accounts) state by setting their state,
// removing the csdb destructed objects and clearing the journal as well as the
// refunds.
func (csdb *CommitStateDB) Finalise(deleteEmptyObjects bool) error {
	for _, dirty := range csdb.journal.dirties {
		stateEntry, exist := csdb.stateObjects[dirty.address]
		if !exist {
			// ripeMD is 'touched' at block 1714175, in tx:
			// 0x1237f737031e40bcde4a8b7e717b2d15e3ecadfe49bb1bbc71ee9deb09c6fcf2
			//
			// That tx goes out of gas, and although the notion of 'touched' does not
			// exist there, the touch-event will still be recorded in the journal.
			// Since ripeMD is a special snowflake, it will persist in the journal even
			// though the journal is reverted. In this special circumstance, it may
			// exist in journal.dirties but not in stateObjects. Thus, we can safely
			// ignore it here.
			continue
		}

		if stateEntry.stateObject.suicided || (deleteEmptyObjects && stateEntry.stateObject.empty()) {
			csdb.deleteStateObject(stateEntry.stateObject)
		} else {
			// Set all the dirty state storage items for the state object in the
			// KVStore and finally set the account in the account mapper.
			stateEntry.stateObject.commitState()
			if err := csdb.updateStateObject(stateEntry.stateObject); err != nil {
				return err
			}
		}

		csdb.stateObjectsDirty[dirty.address] = struct{}{}
	}

	// invalidate journal because reverting across transactions is not allowed
	csdb.clearJournalAndRefund()
	csdb.DeleteLogs()
	return nil
}

// IntermediateRoot returns the current root hash of the state. It is called in
// between transactions to get the root hash that goes into transaction
// receipts.
//
// NOTE: The SDK has not concept or method of getting any intermediate merkle
// root as commitment of the merkle-ized tree doesn't happen until the
// BaseApps' EndBlocker.
func (csdb *CommitStateDB) IntermediateRoot(deleteEmptyObjects bool) (ethcmn.Hash, error) {
	if err := csdb.Finalise(deleteEmptyObjects); err != nil {
		return ethcmn.Hash{}, err
	}

	return ethcmn.Hash{}, nil
}

// updateStateObject writes the given state object to the store.
func (csdb *CommitStateDB) updateStateObject(so *stateObject) error {
	// NOTE: we don't use sdk.NewCoin here to avoid panic on test importer's genesis
	newBalance := sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecFromBigIntWithPrec(so.Balance(), sdk.Precision)} // int2dec
	if !newBalance.IsValid() {
		return fmt.Errorf("invalid balance %s", newBalance)
	}

	//checking and reject tx if address in blacklist
	if csdb.bankKeeper.BlacklistedAddr(so.account.GetAddress()) {
		return fmt.Errorf("address <%s> in blacklist is not allowed", so.account.GetAddress().String())
	}

	coins := so.account.GetCoins()
	balance := coins.AmountOf(newBalance.Denom)
	if balance.IsZero() || !balance.Equal(newBalance.Amount) {
		coins = coins.Add(newBalance)
	}

	if err := so.account.SetCoins(coins); err != nil {
		return err
	}

	csdb.accountKeeper.SetAccount(csdb.ctx, so.account)
	if !csdb.ctx.IsCheckTx() {
		if csdb.Watcher.Enabled() {
			csdb.Watcher.SaveAccount(so.account, false)
		}
	}
	// return csdb.bankKeeper.SetBalance(csdb.ctx, so.account.Address, newBalance)
	return nil
}

// deleteStateObject removes the given state object from the state store.
func (csdb *CommitStateDB) deleteStateObject(so *stateObject) {
	so.deleted = true
	csdb.accountKeeper.RemoveAccount(csdb.ctx, so.account)
}

// ----------------------------------------------------------------------------
// Snapshotting
// ----------------------------------------------------------------------------

// Snapshot returns an identifier for the current revision of the state.
func (csdb *CommitStateDB) Snapshot() int {
	if !csdb.ctx.IsCheckTx() {
		funcName := "Snapshot"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	id := csdb.nextRevisionID
	csdb.nextRevisionID++

	csdb.validRevisions = append(
		csdb.validRevisions,
		revision{
			id:           id,
			journalIndex: csdb.journal.length(),
		},
	)

	return id
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (csdb *CommitStateDB) RevertToSnapshot(revID int) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "RevertToSnapshot"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	// find the snapshot in the stack of valid snapshots
	idx := sort.Search(len(csdb.validRevisions), func(i int) bool {
		return csdb.validRevisions[i].id >= revID
	})

	if idx == len(csdb.validRevisions) || csdb.validRevisions[idx].id != revID {
		panic(fmt.Errorf("revision ID %v cannot be reverted", revID))
	}

	snapshot := csdb.validRevisions[idx].journalIndex

	// replay the journal to undo changes and remove invalidated snapshots
	csdb.journal.revert(csdb, snapshot)
	csdb.validRevisions = csdb.validRevisions[:idx]
}

// ----------------------------------------------------------------------------
// Auxiliary
// ----------------------------------------------------------------------------

// Database retrieves the low level database supporting the lower level trie
// ops. It is not used in Ethermint, so it returns nil.
func (csdb *CommitStateDB) Database() ethstate.Database {
	return nil
}

// Empty returns whether the state object is either non-existent or empty
// according to the EIP161 specification (balance = nonce = code = 0).
func (csdb *CommitStateDB) Empty(addr ethcmn.Address) bool {
	if !csdb.ctx.IsCheckTx() {
		funcName := "Empty"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.getStateObject(addr)
	return so == nil || so.empty()
}

// Exist reports whether the given account address exists in the state. Notably,
// this also returns true for suicided accounts.
func (csdb *CommitStateDB) Exist(addr ethcmn.Address) bool {
	if !csdb.ctx.IsCheckTx() {
		funcName := "Exist"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	return csdb.getStateObject(addr) != nil
}

// Error returns the first non-nil error the StateDB encountered.
func (csdb *CommitStateDB) Error() error {
	return csdb.dbErr
}

// Suicide marks the given account as suicided and clears the account balance.
//
// The account's state object is still available until the state is committed,
// getStateObject will return a non-nil account after Suicide.
func (csdb *CommitStateDB) Suicide(addr ethcmn.Address) bool {
	if !csdb.ctx.IsCheckTx() {
		funcName := "Suicide"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.getStateObject(addr)
	if so == nil {
		return false
	}

	csdb.journal.append(suicideChange{
		account:     &addr,
		prev:        so.suicided,
		prevBalance: sdk.NewDecFromBigIntWithPrec(so.Balance(), sdk.Precision), // int2dec
	})

	so.markSuicided()
	so.SetBalance(new(big.Int))

	return true
}

// Reset clears out all ephemeral state objects from the state db, but keeps
// the underlying account mapper and store keys to avoid reloading data for the
// next operations.
func (csdb *CommitStateDB) Reset(_ ethcmn.Hash) error {
	csdb.stateObjects = make(map[ethcmn.Address]*stateEntry)
	csdb.stateObjectsDirty = make(map[ethcmn.Address]struct{})
	csdb.thash = ethcmn.Hash{}
	csdb.bhash = ethcmn.Hash{}
	csdb.txIndex = 0
	csdb.logSize = 0
	csdb.preimages = []preimageEntry{}
	csdb.hashToPreimageIndex = make(map[ethcmn.Hash]int)
	csdb.accessList = newAccessList()
	csdb.params = nil

	csdb.clearJournalAndRefund()
	return nil
}

// UpdateAccounts updates the nonce and coin balances of accounts
func (csdb *CommitStateDB) UpdateAccounts() {
	for _, stateEntry := range csdb.stateObjects {
		currAcc := csdb.accountKeeper.GetAccount(csdb.ctx, sdk.AccAddress(stateEntry.address.Bytes()))
		ethermintAcc, ok := currAcc.(*ethermint.EthAccount)
		if !ok {
			continue
		}

		balance := sdk.Coin{
			Denom:  sdk.DefaultBondDenom,
			Amount: ethermintAcc.GetCoins().AmountOf(sdk.DefaultBondDenom),
		}

		if stateEntry.stateObject.Balance() != balance.Amount.BigInt() && balance.IsValid() ||
			stateEntry.stateObject.Nonce() != ethermintAcc.GetSequence() {
			stateEntry.stateObject.account = ethermintAcc
		}
	}
}

// ClearStateObjects clears cache of state objects to handle account changes outside of the EVM
func (csdb *CommitStateDB) ClearStateObjects() {
	csdb.stateObjects = make(map[ethcmn.Address]*stateEntry)
	csdb.stateObjectsDirty = make(map[ethcmn.Address]struct{})
}

func (csdb *CommitStateDB) clearJournalAndRefund() {
	csdb.journal = newJournal()
	csdb.validRevisions = csdb.validRevisions[:0]
	csdb.refund = 0
}

// Prepare sets the current transaction hash and index and block hash which is
// used when the EVM emits new state logs.
func (csdb *CommitStateDB) Prepare(thash, bhash ethcmn.Hash, txi int) {
	csdb.thash = thash
	csdb.bhash = bhash
	csdb.txIndex = txi
}

// CreateAccount explicitly creates a state object. If a state object with the
// address already exists the balance is carried over to the new account.
//
// CreateAccount is called during the EVM CREATE operation. The situation might
// arise that a contract does the following:
//
//   1. sends funds to sha(account ++ (nonce + 1))
//   2. tx_create(sha(account ++ nonce)) (note that this gets the address of 1)
//
// Carrying over the balance ensures that Ether doesn't disappear.
func (csdb *CommitStateDB) CreateAccount(addr ethcmn.Address) {
	if !csdb.ctx.IsCheckTx() {
		funcName := "CreateAccount"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	newobj, prevobj := csdb.createObject(addr)
	if prevobj != nil {
		newobj.setBalance(sdk.DefaultBondDenom, sdk.NewDecFromBigIntWithPrec(prevobj.Balance(), sdk.Precision)) // int2dec
	}
}

// ForEachStorage iterates over each storage items, all invoke the provided
// callback on each key, value pair.
func (csdb *CommitStateDB) ForEachStorage(addr ethcmn.Address, cb func(key, value ethcmn.Hash) (stop bool)) error {
	if !csdb.ctx.IsCheckTx() {
		funcName := "ForEachStorage"
		analyzer.StartTxLog(funcName)
		defer analyzer.StopTxLog(funcName)
	}

	so := csdb.getStateObject(addr)
	if so == nil {
		return nil
	}

	store := csdb.ctx.KVStore(csdb.storeKey)
	prefix := AddressStoragePrefix(so.Address())
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := ethcmn.BytesToHash(iterator.Key())
		value := ethcmn.BytesToHash(iterator.Value())

		if idx, dirty := so.keyToDirtyStorageIndex[key]; dirty {
			// check if iteration stops
			if cb(key, so.dirtyStorage[idx].Value) {
				break
			}

			continue
		}

		// check if iteration stops
		if cb(key, value) {
			return nil
		}
	}

	return nil
}

// GetOrNewStateObject retrieves a state object or create a new state object if
// nil.
func (csdb *CommitStateDB) GetOrNewStateObject(addr ethcmn.Address) StateObject {
	so := csdb.getStateObject(addr)
	if so == nil || so.deleted {
		so, _ = csdb.createObject(addr)
	}

	return so
}

// createObject creates a new state object. If there is an existing account with
// the given address, it is overwritten and returned as the second return value.
func (csdb *CommitStateDB) createObject(addr ethcmn.Address) (newObj, prevObj *stateObject) {
	prevObj = csdb.getStateObject(addr)

	acc := csdb.accountKeeper.NewAccountWithAddress(csdb.ctx, sdk.AccAddress(addr.Bytes()))

	newObj = newStateObject(csdb, acc)
	newObj.setNonce(0) // sets the object to dirty

	if prevObj == nil {
		csdb.journal.append(createObjectChange{account: &addr})
	} else {
		csdb.journal.append(resetObjectChange{prev: prevObj})
	}

	csdb.setStateObject(newObj)
	return newObj, prevObj
}

// setError remembers the first non-nil error it is called with.
func (csdb *CommitStateDB) setError(err error) {
	if csdb.dbErr == nil {
		csdb.dbErr = err
	}
}

// getStateObject attempts to retrieve a state object given by the address.
// Returns nil and sets an error if not found.
func (csdb *CommitStateDB) getStateObject(addr ethcmn.Address) (stateObject *stateObject) {
	if v, found := csdb.stateObjects[addr]; found {
		// prefer 'live' (cached) objects
		if so := v.stateObject; so != nil {
			if so.deleted {
				return nil
			}

			return so
		}
	}

	// otherwise, attempt to fetch the account from the account mapper
	acc := csdb.accountKeeper.GetAccount(csdb.ctx, sdk.AccAddress(addr.Bytes()))
	if acc == nil {
		csdb.setError(fmt.Errorf("no account found for address: %s", EthAddressStringer(addr).String()))
		return nil
	}

	// insert the state object into the live set
	so := newStateObject(csdb, acc)
	csdb.setStateObject(so)

	return so
}

func (csdb *CommitStateDB) setStateObject(so *stateObject) {
	if _, found := csdb.stateObjects[so.Address()]; found {
		// update the existing object
		csdb.stateObjects[so.Address()].stateObject = so
		return
	}

	// append the new state object to the stateObjects slice
	se := &stateEntry{
		address:     so.Address(),
		stateObject: so,
	}

	csdb.stateObjects[se.address] = se
}

// RawDump returns a raw state dump.
//
// TODO: Implement if we need it, especially for the RPC API.
func (csdb *CommitStateDB) RawDump() ethstate.Dump {
	return ethstate.Dump{}
}

type preimageEntry struct {
	// hash key of the preimage entry
	hash     ethcmn.Hash
	preimage []byte
}

func (csdb *CommitStateDB) SetLogSize(logSize uint) {
	csdb.logSize = logSize
}

func (csdb *CommitStateDB) GetLogSize() uint {
	return csdb.logSize
}

// SetContractDeploymentWhitelistMember sets the target address list into whitelist store
func (csdb *CommitStateDB) SetContractDeploymentWhitelist(addrList AddressList) {
	if csdb.Watcher.Enabled() {
		for i := 0; i < len(addrList); i++ {
			csdb.Watcher.SaveContractDeploymentWhitelistItem(addrList[i])
		}
	}
	store := csdb.ctx.KVStore(csdb.storeKey)
	for i := 0; i < len(addrList); i++ {
		store.Set(GetContractDeploymentWhitelistMemberKey(addrList[i]), []byte(""))
	}
}

// DeleteContractDeploymentWhitelist deletes the target address list from whitelist store
func (csdb *CommitStateDB) DeleteContractDeploymentWhitelist(addrList AddressList) {
	if csdb.Watcher.Enabled() {
		for i := 0; i < len(addrList); i++ {
			csdb.Watcher.DeleteContractDeploymentWhitelist(addrList[i])
		}
	}
	store := csdb.ctx.KVStore(csdb.storeKey)
	for i := 0; i < len(addrList); i++ {
		store.Delete(GetContractDeploymentWhitelistMemberKey(addrList[i]))
	}
}

// GetContractDeploymentWhitelist gets the whole contract deployment whitelist currently
func (csdb *CommitStateDB) GetContractDeploymentWhitelist() (whitelist AddressList) {
	store := csdb.ctx.KVStore(csdb.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, KeyPrefixContractDeploymentWhitelist)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		whitelist = append(whitelist, splitApprovedDeployerAddress(iterator.Key()))
	}

	return
}

// IsDeployerInWhitelist checks whether the deployer is in the whitelist as a distributor
func (csdb *CommitStateDB) IsDeployerInWhitelist(deployerAddr sdk.AccAddress) bool {
	bs := csdb.dbAdapter.NewStore(csdb.ctx.KVStore(csdb.storeKey), KeyPrefixContractDeploymentWhitelist)
	return bs.Has(deployerAddr)
}

// SetContractBlockedList sets the target address list into blocked list store
func (csdb *CommitStateDB) SetContractBlockedList(addrList AddressList) {
	defer GetEvmParamsCache().SetNeedBlockedUpdate()
	if csdb.Watcher.Enabled() {
		for i := 0; i < len(addrList); i++ {
			csdb.Watcher.SaveContractBlockedListItem(addrList[i])
		}
	}
	store := csdb.ctx.KVStore(csdb.storeKey)
	for i := 0; i < len(addrList); i++ {
		store.Set(GetContractBlockedListMemberKey(addrList[i]), []byte(""))
	}
}

// DeleteContractBlockedList deletes the target address list from blocked list store
func (csdb *CommitStateDB) DeleteContractBlockedList(addrList AddressList) {
	defer GetEvmParamsCache().SetNeedBlockedUpdate()
	if csdb.Watcher.Enabled() {
		for i := 0; i < len(addrList); i++ {
			csdb.Watcher.DeleteContractBlockedList(addrList[i])
		}
	}
	store := csdb.ctx.KVStore(csdb.storeKey)
	for i := 0; i < len(addrList); i++ {
		store.Delete(GetContractBlockedListMemberKey(addrList[i]))
	}
}

// GetContractBlockedList gets the whole contract blocked list currently
func (csdb *CommitStateDB) GetContractBlockedList() (blockedList AddressList) {
	store := csdb.ctx.KVStore(csdb.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, KeyPrefixContractBlockedList)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		if len(iterator.Value()) == 0 {
			blockedList = append(blockedList, splitBlockedContractAddress(iterator.Key()))
		}
	}

	return
}

// IsContractInBlockedList checks whether the contract address is in the blocked list
func (csdb *CommitStateDB) IsContractInBlockedList(contractAddr sdk.AccAddress) bool {
	bc := csdb.GetContractMethodBlockedByAddress(contractAddr)
	if bc == nil {
		//contractAddr is not blocked
		return false
	}
	// check contractAddr whether block full-method and special-method
	return bc.IsAllMethodBlocked()
}

// GetContractMethodBlockedByAddress gets contract methods blocked by address
func (csdb *CommitStateDB) GetContractMethodBlockedByAddress(contractAddr sdk.AccAddress) *BlockedContract {
	if csdb.ctx.IsDeliver() {
		if GetEvmParamsCache().IsNeedBlockedUpdate() {
			bcl := csdb.GetContractMethodBlockedList()
			GetEvmParamsCache().UpdateBlockedContractMethod(bcl)
		}
		return GetEvmParamsCache().GetBlockedContractMethod(contractAddr.String())
	}

	//use dbAdapter for watchdb or prefixdb
	bs := csdb.dbAdapter.NewStore(csdb.ctx.KVStore(csdb.storeKey), KeyPrefixContractBlockedList)
	vaule := bs.Get(contractAddr)
	if vaule == nil {
		// address is not exist
		return nil
	} else {
		methods := ContractMethods{}
		var bc *BlockedContract
		if len(vaule) == 0 {
			//address is exist,but the blocked is old version.
			bc = NewBlockContract(contractAddr, methods)
		} else {
			// get block contract from cache without anmio
			if contractMethodBlockedCache != nil {
				if cm, ok := contractMethodBlockedCache.GetContractMethod(vaule); ok {
					return NewBlockContract(contractAddr, cm)
				}
			}
			//address is exist,but the blocked is new version.
			csdb.cdc.MustUnmarshalJSON(vaule, &methods)
			bc = NewBlockContract(contractAddr, methods)

			// write block contract into cache
			if contractMethodBlockedCache != nil {
				contractMethodBlockedCache.SetContractMethod(vaule, methods)
			}
		}
		return bc
	}
}

// InsertContractMethodBlockedList sets the list of contract method blocked into blocked list store
func (csdb *CommitStateDB) InsertContractMethodBlockedList(contractList BlockedContractList) sdk.Error {
	defer GetEvmParamsCache().SetNeedBlockedUpdate()
	for i := 0; i < len(contractList); i++ {
		bc := csdb.GetContractMethodBlockedByAddress(contractList[i].Address)
		if bc != nil {
			result, err := bc.BlockMethods.InsertContractMethods(contractList[i].BlockMethods)
			if err != nil {
				return err
			}
			bc.BlockMethods = result
		} else {
			bc = &contractList[i]
		}

		csdb.SetContractMethodBlocked(*bc)
	}
	return nil
}

// DeleteContractMethodBlockedList delete the list of contract method blocked  from blocked list store
func (csdb *CommitStateDB) DeleteContractMethodBlockedList(contractList BlockedContractList) sdk.Error {
	defer GetEvmParamsCache().SetNeedBlockedUpdate()
	for i := 0; i < len(contractList); i++ {
		bc := csdb.GetContractMethodBlockedByAddress(contractList[i].Address)
		if bc != nil {
			result, err := bc.BlockMethods.DeleteContractMethodMap(contractList[i].BlockMethods)
			if err != nil {
				return ErrBlockedContractMethodIsNotExist(contractList[i].Address, err)
			}
			bc.BlockMethods = result
			//if block contract method delete empty then remove contract from blocklist.
			if len(bc.BlockMethods) == 0 {
				addressList := AddressList{}
				addressList = append(addressList, bc.Address)
				//in watchdb contract blocked and contract method blocked use same prefix
				//so delete contract method blocked is can use function of delete contract blocked
				csdb.DeleteContractBlockedList(addressList)
			} else {
				csdb.SetContractMethodBlocked(*bc)
			}
		} else {
			return ErrBlockedContractMethodIsNotExist(contractList[i].Address, ErrorContractMethodBlockedIsNotExist)
		}
	}
	return nil
}

// GetContractMethodBlockedList get the list of contract method blocked from blocked list store
func (csdb *CommitStateDB) GetContractMethodBlockedList() (blockedContractList BlockedContractList) {
	store := csdb.ctx.KVStore(csdb.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, KeyPrefixContractBlockedList)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		addr := sdk.AccAddress(splitBlockedContractAddress(iterator.Key()))
		value := iterator.Value()
		methods := ContractMethods{}
		if len(value) != 0 {
			csdb.cdc.MustUnmarshalJSON(value, &methods)
		}
		bc := NewBlockContract(addr, methods)
		blockedContractList = append(blockedContractList, *bc)
	}
	return
}

// IsContractMethodBlocked checks whether the contract method is blocked
func (csdb *CommitStateDB) IsContractMethodBlocked(contractAddr sdk.AccAddress, method string) bool {
	bc := csdb.GetContractMethodBlockedByAddress(contractAddr)
	if bc == nil {
		//contractAddr is not blocked
		return false
	}
	// it maybe happens,because ok_verifier verify called before getCode, for executing old logic follow code is return false
	if bc.IsAllMethodBlocked() {
		return false
	}
	// check contractAddr whether block full-method and special-method
	return bc.IsMethodBlocked(method)
}

// SetContractMethodBlocked sets contract method blocked into blocked list store
func (csdb *CommitStateDB) SetContractMethodBlocked(contract BlockedContract) {
	key := GetContractBlockedListMemberKey(contract.Address)
	SortContractMethods(contract.BlockMethods)
	value := csdb.cdc.MustMarshalJSON(contract.BlockMethods)
	value = sdk.MustSortJSON(value)
	store := csdb.ctx.KVStore(csdb.storeKey)
	store.Set(key, value)

	watcherEnable := csdb.Watcher.Enabled()
	if watcherEnable {
		csdb.Watcher.SaveContractMethodBlockedListItem(contract.Address, value)
	}
}
