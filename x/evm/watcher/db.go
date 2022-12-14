package watcher

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/FiboChain/fbc/libs/cosmos-sdk/client/flags"
	dbm "github.com/FiboChain/fbc/libs/tm-db"
	"github.com/spf13/viper"
)

const (
	FlagFastQuery    = "fast-query"
	FlagFastQueryLru = "fast-lru"
	FlagDBBackend    = "db_backend"
	FlagCheckWd      = "check_watchdb"

	WatchDbDir  = "data"
	WatchDBName = "watch"
)

type WatchStore struct {
	db dbm.DB
}

var gWatchStore *WatchStore = nil
var once sync.Once

func InstanceOfWatchStore() *WatchStore {
	once.Do(func() {
		if IsWatcherEnabled() {
			gWatchStore = &WatchStore{db: initDb()}
		}
	})
	return gWatchStore
}

func initDb() dbm.DB {
	homeDir := viper.GetString(flags.FlagHome)
	dbPath := filepath.Join(homeDir, WatchDbDir)
	backend := viper.GetString(FlagDBBackend)
	if backend == "" {
		backend = string(dbm.GoLevelDBBackend)
	}

	return dbm.NewDB(WatchDBName, dbm.BackendType(backend), dbPath)
}

func (w WatchStore) Set(key []byte, value []byte) {
	err := w.db.Set(key, value)
	if err != nil {
		log.Println("watchdb error: ", err.Error())
	}
}

func (w WatchStore) Get(key []byte) ([]byte, error) {
	return w.db.Get(key)
}

func (w WatchStore) GetUnsafe(key []byte, processor dbm.UnsafeValueProcessor) (interface{}, error) {
	return w.db.GetUnsafeValue(key, processor)
}

func (w WatchStore) Delete(key []byte) {
	err := w.db.Delete(key)
	if err != nil {
		log.Printf("watchdb error: " + err.Error())
	}
}

func (w WatchStore) Has(key []byte) bool {
	res, err := w.db.Has(key)
	if err != nil {
		log.Println("watchdb error: " + err.Error())
		return false
	}
	return res
}

func (w WatchStore) Iterator(start, end []byte) dbm.Iterator {
	it, err := w.db.Iterator(start, end)
	if err != nil {
		log.Println("watchdb error: " + err.Error())
		return nil
	}
	return it
}
