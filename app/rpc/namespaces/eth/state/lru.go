package state

import (
	"errors"
	"sync"

	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"

	lru "github.com/hashicorp/golang-lru"
)

//the default lru cache size is 1kw, that means the max memory size we needs is (32 + 32 + 4) * 10000000, about 700MB
var (
	defaultLruSize int = 10000000
	gStateLru      *lru.Cache
	once           sync.Once
)

//redefine fast-query to avoid cycle package import
const FlagFastQuery = "fast-query"

func isWatcherEnabled() bool {
	return viper.GetBool(FlagFastQuery)
}

func InstanceOfStateLru() *lru.Cache {
	once.Do(func() {
		if isWatcherEnabled() {
			var e error = nil
			gStateLru, e = lru.New(defaultLruSize)
			if e != nil {
				panic(errors.New("Failed to call InstanceOfStateLru cause :" + e.Error()))
			}
		}
	})
	return gStateLru
}

func GetStateFromLru(key []byte) []byte {
	cache := InstanceOfStateLru()
	if cache == nil {
		return nil
	}
	value, ok := cache.Get(amino.BytesToStr(key))
	if ok {
		ret, ok := value.([]byte)
		if ok {
			return ret
		}
	}
	return nil
}

func SetStateToLru(key []byte, value []byte) {
	cache := InstanceOfStateLru()
	if cache == nil {
		return
	}
	cache.Add(amino.BytesToStr(key), value)
}
