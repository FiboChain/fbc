package types

import (
	"fmt"
	"sync"
	"time"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/FiboChain/fbc/libs/cosmos-sdk/store/types"
	"github.com/FiboChain/fbc/libs/tendermint/crypto"
	"github.com/FiboChain/fbc/libs/tendermint/libs/log"
	"github.com/spf13/viper"
)

var (
	maxAccInMap        = 100000
	deleteAccCount     = 10000
	maxStorageInMap    = 10000000
	deleteStorageCount = 1000000

	FlagMultiCache         = "multi-cache"
	MaxAccInMultiCache     = "multi-cache-acc"
	MaxStorageInMultiCache = "multi-cache-storage"
	UseCache               bool
)

type Account interface {
	Copy() Account
	GetAddress() AccAddress
	SetAddress(AccAddress) error
	GetPubKey() crypto.PubKey
	SetPubKey(crypto.PubKey) error
	GetAccountNumber() uint64
	SetAccountNumber(uint64) error
	GetSequence() uint64
	SetSequence(uint64) error
	GetCoins() Coins
	SetCoins(Coins) error
	SpendableCoins(blockTime time.Time) Coins
	String() string
}

type ModuleAccount interface {
	Account

	GetName() string
	GetPermissions() []string
	HasPermission(string) bool
}

type storageWithCache struct {
	value []byte
	dirty bool
}

type accountWithCache struct {
	acc     Account
	gas     uint64
	isDirty bool
}

type codeWithCache struct {
	code    []byte
	isDirty bool
}

type Cache struct {
	useCache  bool
	parent    *Cache
	gasConfig types.GasConfig

	storageMap map[ethcmn.Address]map[ethcmn.Hash]*storageWithCache
	accMap     map[ethcmn.Address]*accountWithCache
	codeMap    map[ethcmn.Hash]*codeWithCache

	accMutex sync.RWMutex
}

func initCacheParam() {
	UseCache = false

	if data := viper.GetInt(MaxAccInMultiCache); data != 0 {
		maxAccInMap = data
		deleteAccCount = maxAccInMap / 10
	}

	if data := viper.GetInt(MaxStorageInMultiCache); data != 0 {
		maxStorageInMap = data
		deleteStorageCount = maxStorageInMap / 10
	}
}

func NewChainCache() *Cache {
	initCacheParam()
	return NewCache(nil, UseCache)
}

func NewCache(parent *Cache, useCache bool) *Cache {
	return &Cache{
		useCache: useCache,
		parent:   parent,

		storageMap: make(map[ethcmn.Address]map[ethcmn.Hash]*storageWithCache),
		accMap:     make(map[ethcmn.Address]*accountWithCache),
		codeMap:    make(map[ethcmn.Hash]*codeWithCache),
		gasConfig:  types.KVGasConfig(),
	}

}

func (c *Cache) skip() bool {
	if c == nil || !c.useCache {
		return true
	}
	return false
}

func (c *Cache) IsEnabled() bool {
	return !c.skip()
}

func (c *Cache) DisableCache() {
	c.useCache = false
}

func (c *Cache) UpdateAccount(addr AccAddress, acc Account, lenBytes int, isDirty bool) {
	if c.skip() {
		return
	}
	ethAddr := ethcmn.BytesToAddress(addr.Bytes())
	accWithCache := &accountWithCache{
		acc:     acc,
		isDirty: isDirty,
		gas:     types.Gas(lenBytes)*c.gasConfig.ReadCostPerByte + c.gasConfig.ReadCostFlat,
	}
	c.accMutex.Lock()
	c.accMap[ethAddr] = accWithCache
	c.accMutex.Unlock()
}

func (c *Cache) UpdateStorage(addr ethcmn.Address, key ethcmn.Hash, value []byte, isDirty bool) {
	if c.skip() {
		return
	}

	if _, ok := c.storageMap[addr]; !ok {
		c.storageMap[addr] = make(map[ethcmn.Hash]*storageWithCache, 0)
	}
	c.storageMap[addr][key] = &storageWithCache{
		value: value,
		dirty: isDirty,
	}
}

func (c *Cache) UpdateCode(key []byte, value []byte, isdirty bool) {
	if c.skip() {
		return
	}
	hash := ethcmn.BytesToHash(key)
	c.codeMap[hash] = &codeWithCache{
		code:    value,
		isDirty: isdirty,
	}
}

func (c *Cache) GetAccount(addr ethcmn.Address) (Account, uint64, bool) {
	if c.skip() {
		return nil, 0, false
	}

	c.accMutex.RLock()
	data, ok := c.accMap[addr]
	c.accMutex.RUnlock()
	if ok {
		return data.acc, data.gas, ok
	}

	if c.parent != nil {
		acc, gas, ok := c.parent.GetAccount(addr)
		return acc, gas, ok
	}
	return nil, 0, false
}

func (c *Cache) GetStorage(addr ethcmn.Address, key ethcmn.Hash) ([]byte, bool) {
	if c.skip() {
		return nil, false
	}
	if _, hasAddr := c.storageMap[addr]; hasAddr {
		data, hasKey := c.storageMap[addr][key]
		if hasKey {
			return data.value, hasKey
		}
	}

	if c.parent != nil {
		return c.parent.GetStorage(addr, key)
	}
	return nil, false
}

func (c *Cache) GetCode(key []byte) ([]byte, bool) {
	if c.skip() {
		return nil, false
	}

	hash := ethcmn.BytesToHash(key)
	if data, ok := c.codeMap[hash]; ok {
		return data.code, ok
	}

	if c.parent != nil {
		return c.parent.GetCode(hash.Bytes())
	}
	return nil, false
}

func (c *Cache) Write(updateDirty bool) {
	if c.skip() {
		return
	}

	if c.parent == nil {
		return
	}

	c.writeStorage(updateDirty)
	c.writeAcc(updateDirty)
	c.writeCode(updateDirty)
}

func (c *Cache) writeStorage(updateDirty bool) {
	for addr, storages := range c.storageMap {
		if _, ok := c.parent.storageMap[addr]; !ok {
			c.parent.storageMap[addr] = make(map[ethcmn.Hash]*storageWithCache, 0)
		}

		for key, v := range storages {
			if needWriteToParent(updateDirty, v.dirty) {
				c.parent.storageMap[addr][key] = v
			}
		}
	}
	c.storageMap = make(map[ethcmn.Address]map[ethcmn.Hash]*storageWithCache)
}

func (c *Cache) setAcc(addr ethcmn.Address, v *accountWithCache) {
	c.accMutex.Lock()
	c.accMap[addr] = v
	c.accMutex.Unlock()
}

func (c *Cache) writeAcc(updateDirty bool) {
	c.accMutex.RLock()
	for addr, v := range c.accMap {
		if needWriteToParent(updateDirty, v.isDirty) {
			c.parent.setAcc(addr, v)
		}
	}
	c.accMutex.RUnlock()

	c.accMutex.Lock()
	for k := range c.accMap {
		delete(c.accMap, k)
	}
	c.accMutex.Unlock()
}

func (c *Cache) writeCode(updateDirty bool) {
	for hash, v := range c.codeMap {
		if needWriteToParent(updateDirty, v.isDirty) {
			c.parent.codeMap[hash] = v
		}
	}
	c.codeMap = make(map[ethcmn.Hash]*codeWithCache)
}

func needWriteToParent(updateDirty bool, dirty bool) bool {
	// not dirty
	if !dirty {
		return true
	}

	// dirty
	if updateDirty {
		return true
	}
	return false
}

func (c *Cache) storageSize() int {
	lenStorage := 0
	for _, v := range c.storageMap {
		lenStorage += len(v)
	}
	return lenStorage
}

func (c *Cache) TryDelete(logger log.Logger, height int64) {
	if c.skip() {
		return
	}
	if height%1000 == 0 {
		c.logInfo(logger, "null")
	}

	lenStorage := c.storageSize()
	if c.lenOfAccMap() < maxAccInMap && lenStorage < maxStorageInMap {
		return
	}

	deleteMsg := ""
	lenOfAcc := c.lenOfAccMap()
	if lenOfAcc >= maxAccInMap {
		deleteMsg += fmt.Sprintf("Acc:Deleted Before:%d", lenOfAcc)
		cnt := 0
		c.accMutex.Lock()
		for key := range c.accMap {
			delete(c.accMap, key)
			cnt++
			if cnt > deleteAccCount {
				break
			}
		}
		c.accMutex.Unlock()
	}

	if lenStorage >= maxStorageInMap {
		deleteMsg += fmt.Sprintf("Storage:Deleted Before:len(contract):%d, len(storage):%d", len(c.storageMap), lenStorage)
		cnt := 0
		for key, value := range c.storageMap {
			cnt += len(value)
			delete(c.storageMap, key)
			if cnt > deleteStorageCount {
				break
			}
		}
	}
	if deleteMsg != "" {
		c.logInfo(logger, deleteMsg)
	}
}

func (c *Cache) logInfo(logger log.Logger, deleteMsg string) {
	nowStats := fmt.Sprintf("len(acc):%d len(contracts):%d len(storage):%d", c.lenOfAccMap(), len(c.storageMap), c.storageSize())
	logger.Info("MultiCache", "deleteMsg", deleteMsg, "nowStats", nowStats)
}

func (c *Cache) lenOfAccMap() (l int) {
	c.accMutex.RLock()
	l = len(c.accMap)
	c.accMutex.RUnlock()
	return
}
