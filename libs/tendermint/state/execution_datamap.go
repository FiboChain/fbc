package state

import (
	"container/list"
	"sync"
)

type deltaMap struct {
	mtx       sync.Mutex
	cacheMap  map[int64]*list.Element
	cacheList *list.List
	mrh       int64
}

func newDataMap() *deltaMap {
	return &deltaMap{
		cacheMap:  make(map[int64]*list.Element),
		cacheList: list.New(),
	}
}

type payload struct {
	h  int64
	di *DeltaInfo
}

func (m *deltaMap) insert(height int64, deltaInfo *DeltaInfo, mrh int64) {
	if deltaInfo == nil {
		return
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()
	e := m.cacheList.PushBack(&payload{height, deltaInfo})
	m.cacheMap[height] = e
	m.mrh = mrh
}

func (m *deltaMap) fetch(height int64) (*DeltaInfo, int64) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	popped := m.cacheMap[height]
	delete(m.cacheMap, height)
	if popped != nil {
		m.cacheList.Remove(popped)
		pl := popped.Value.(*payload)
		return pl.di, m.mrh
	}

	return nil, m.mrh
}

// remove all elements no higher than target
func (m *deltaMap) remove(target int64) (int, int) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	num := 0
	for {
		e := m.cacheList.Front()
		if e == nil {
			break
		}
		h := e.Value.(*payload).h
		if h > target {
			break
		}
		m.cacheList.Remove(e)
		delete(m.cacheMap, h)
		num++
	}

	return num, len(m.cacheMap)
}

func (m *deltaMap) info() (int, int) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return len(m.cacheMap), m.cacheList.Len()
}
