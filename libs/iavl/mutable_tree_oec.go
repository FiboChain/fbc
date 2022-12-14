package iavl

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/FiboChain/fbc/libs/iavl/trace"

	dbm "github.com/FiboChain/fbc/libs/tm-db"
)

const (
	minHistoryStateNum             = 30
	FlagIavlCommitIntervalHeight   = "iavl-commit-interval-height"
	FlagIavlMinCommitItemCount     = "iavl-min-commit-item-count"
	FlagIavlHeightOrphansCacheSize = "iavl-height-orphans-cache-size"
	FlagIavlMaxCommittedHeightNum  = "iavl-max-committed-height-num"
	FlagIavlEnableAsyncCommit      = "iavl-enable-async-commit"
)

var (
	// ErrVersionDoesNotExist is returned if a requested version does not exist.
	ErrVersionDoesNotExist = errors.New("version does not exist")

	// Parameters below here are changed from cosmos-sdk, controlled by flag
	CommitIntervalHeight      int64 = 100
	MinCommitItemCount        int64 = 500000
	HeightOrphansCacheSize          = 8
	MaxCommittedHeightNum           = minHistoryStateNum
	EnableAsyncCommit               = false
	EnablePruningHistoryState       = true
)

type commitEvent struct {
	version    int64
	versions   map[int64]bool
	batch      dbm.Batch
	tpp        map[string]*Node
	wg         *sync.WaitGroup
	iavlHeight int
}

func (tree *MutableTree) SaveVersionAsync(version int64, useDeltas bool) ([]byte, int64, error) {
	moduleName := tree.GetModuleName()
	oldRoot, saved := tree.hasSaved(version)
	if saved {
		return nil, version, fmt.Errorf("existing version: %d, root: %X", version, oldRoot)
	}

	batch := tree.NewBatch()
	if tree.root != nil {
		if useDeltas {
			tree.updateBranchWithDelta(tree.root)
		} else if produceDelta {
			tree.ndb.updateBranchConcurrency(tree.root, tree.savedNodes)
		} else {
			tree.ndb.updateBranchConcurrency(tree.root, nil)
		}

		// generate state delta
		if produceDelta {
			delete(tree.savedNodes, string(tree.root.hash))
			tree.savedNodes["root"] = tree.root
			tree.GetDelta()
		}
	}

	tree.ndb.SaveOrphans(batch, version, tree.orphans)

	shouldPersist := (version-tree.lastPersistHeight >= CommitIntervalHeight) ||
		(treeMap.totalPreCommitCacheSize >= MinCommitItemCount)

	if shouldPersist {
		if err := tree.persist(batch, version); err != nil {
			return nil, 0, err
		}
	} else {
		batch.Close()
	}

	// set new working tree
	tree.ImmutableTree = tree.ImmutableTree.clone()
	tree.lastSaved = tree.ImmutableTree.clone()
	tree.orphans = []*Node{}
	for k := range tree.savedNodes {
		delete(tree.savedNodes, k)
	}

	rootHash := tree.lastSaved.Hash()
	tree.setHeightOrphansItem(version, rootHash)

	tree.version = version
	if shouldPersist {
		tree.versions.Set(version, true)
	}
	treeMap.updateMutableTreeMap(moduleName)

	tree.removedVersions.Range(func(k, v interface{}) bool {
		tree.log(IavlDebug, "remove version from tree version map", "Height", k.(int64))
		tree.removeVersion(k.(int64))
		tree.removedVersions.Delete(k)
		return true
	})

	tree.ndb.log(IavlDebug, tree.ndb.sprintCacheLog(version))
	return rootHash, version, nil
}

func (tree *MutableTree) removeVersion(version int64) {
	tree.versions.Delete(version)
}

func (tree *MutableTree) persist(batch dbm.Batch, version int64) error {
	tree.commitCh <- commitEvent{-1, nil, nil, nil, nil, 0}
	var tpp map[string]*Node = nil
	if EnablePruningHistoryState {
		tree.ndb.saveCommitOrphans(batch, version, tree.commitOrphans)
	}
	if tree.root == nil {
		// There can still be orphans, for example if the root is the node being removed.
		if err := tree.ndb.SaveEmptyRoot(batch, version); err != nil {
			return err
		}
	} else {
		if err := tree.ndb.SaveRoot(batch, tree.root, version); err != nil {
			return err
		}
		tpp = tree.ndb.asyncPersistTppStart(version)
	}
	tree.commitOrphans = map[string]int64{}
	versions := tree.deepCopyVersions()
	tree.commitCh <- commitEvent{version, versions, batch,
		tpp, nil, int(tree.Height())}
	tree.lastPersistHeight = version
	return nil
}

func (tree *MutableTree) commitSchedule() {
	tree.loadVersionToCommittedHeightMap()
	for event := range tree.commitCh {
		if event.version < 0 {
			continue
		}
		_, ok := tree.committedHeightMap[event.version]
		if ok {
			if event.wg != nil {
				event.wg.Done()
				break
			}
			continue
		}

		trc := trace.NewTracer()
		trc.Pin("Pruning")
		tree.updateCommittedStateHeightPool(event.batch, event.version, event.versions)

		tree.ndb.persistTpp(&event, trc)

		if event.wg != nil {
			event.wg.Done()
			break
		}
	}
}

func (tree *MutableTree) loadVersionToCommittedHeightMap() {
	versions, err := tree.ndb.getRoots()
	if err != nil {
		tree.log(IavlErr, "failed to get versions from db", "error", err.Error())
	}
	versionSlice := make([]int64, 0, len(versions))
	for version := range versions {
		versionSlice = append(versionSlice, version)
	}
	sort.Slice(versionSlice, func(i, j int) bool {
		return versionSlice[i] < versionSlice[j]
	})
	for _, version := range versionSlice {
		tree.committedHeightMap[version] = true
		tree.committedHeightQueue.PushBack(version)
	}
	if len(versionSlice) > 0 {
		tree.log(IavlInfo, "", "Tree", tree.GetModuleName(), "committed height queue", versionSlice)
	}
}

func (tree *MutableTree) StopTree() {
	tree.log(IavlInfo, "stopping iavl", "commit height", tree.version)
	defer tree.log(IavlInfo, "stopping iavl completed", "commit height", tree.version)

	if !EnableAsyncCommit {
		return
	}

	batch := tree.NewBatch()
	if tree.root == nil {
		if err := tree.ndb.SaveEmptyRoot(batch, tree.version); err != nil {
			panic(err)
		}
	} else {
		if err := tree.ndb.SaveRoot(batch, tree.root, tree.version); err != nil {
			panic(err)
		}
	}
	tpp := tree.ndb.asyncPersistTppStart(tree.version)

	var wg sync.WaitGroup
	wg.Add(1)
	versions := tree.deepCopyVersions()

	tree.commitCh <- commitEvent{tree.version, versions, batch, tpp, &wg, 0}
	wg.Wait()
}

func (tree *MutableTree) log(level int, msg string, kvs ...interface{}) {
	iavlLog(tree.GetModuleName(), level, msg, kvs...)
}

func (tree *MutableTree) setHeightOrphansItem(version int64, rootHash []byte) {
	tree.ndb.setHeightOrphansItem(version, rootHash)
}

func (tree *MutableTree) updateCommittedStateHeightPool(batch dbm.Batch, version int64, versions map[int64]bool) {
	queue := tree.committedHeightQueue
	queue.PushBack(version)
	tree.committedHeightMap[version] = true

	if queue.Len() > tree.historyStateNum {
		item := queue.Front()
		oldVersion := queue.Remove(item).(int64)
		delete(tree.committedHeightMap, oldVersion)

		if EnablePruningHistoryState {

			if err := tree.deleteVersion(batch, oldVersion, versions); err != nil {
				tree.log(IavlErr, "Failed to delete", "height", oldVersion, "error", err.Error())
			} else {
				tree.log(IavlDebug, "History state removed", "version", oldVersion)
				tree.removedVersions.Store(oldVersion, nil)
			}
		}
	}
}

func (tree *MutableTree) GetDBReadTime() int {
	return tree.ndb.getDBReadTime()
}

func (tree *MutableTree) GetDBReadCount() int {
	return tree.ndb.getDBReadCount()
}

func (tree *MutableTree) GetDBWriteCount() int {
	return tree.ndb.getDBWriteCount()
}

func (tree *MutableTree) GetNodeReadCount() int {
	return tree.ndb.getNodeReadCount()
}

func (tree *MutableTree) ResetCount() {
	tree.ndb.resetCount()
}

func (tree *MutableTree) GetModuleName() string {
	return tree.ndb.name
}

func (tree *MutableTree) NewBatch() dbm.Batch {
	return tree.ndb.NewBatch()
}

func (tree *MutableTree) addOrphansOptimized(orphans []*Node) {
	for _, node := range orphans {
		if node.persisted || node.prePersisted {
			if len(node.hash) == 0 {
				panic("Expected to find node hash, but was empty")
			}
			tree.orphans = append(tree.orphans, node)
			if node.persisted && EnablePruningHistoryState {
				k := string(node.hash)
				tree.commitOrphans[k] = node.version
				if produceDelta {
					commitOrp := &CommitOrphansImp{Key: k, CommitValue: node.version}
					tree.deltas.CommitOrphansDelta = append(tree.deltas.CommitOrphansDelta, commitOrp)
				}
			}
		}

	}
}

func (tree *MutableTree) hasSaved(version int64) ([]byte, bool) {
	return tree.ndb.inVersionCacheMap(version)
}

func (tree *MutableTree) deepCopyVersions() map[int64]bool {
	if !EnablePruningHistoryState {
		return nil
	}

	return tree.versions.Clone()
}

func (tree *MutableTree) updateBranchWithDelta(node *Node) []byte {
	node.persisted = false
	node.prePersisted = false

	if node.leftHash != nil {
		key := string(node.leftHash)
		if tmp := tree.savedNodes[key]; tmp != nil {
			node.leftHash = tree.updateBranchWithDelta(tree.savedNodes[key])
		}
	}
	if node.rightHash != nil {
		key := string(node.rightHash)
		if tmp := tree.savedNodes[key]; tmp != nil {
			node.rightHash = tree.updateBranchWithDelta(tree.savedNodes[key])
		}
	}

	node._hash()
	tree.ndb.saveNodeToPrePersistCache(node)

	node.leftNode = nil
	node.rightNode = nil

	// TODO: handle magic number
	tree.savedNodes[string(node.hash)] = node

	return node.hash
}
func (t *ImmutableTree) GetPersistedRoots() map[int64][]byte {
	return t.ndb.roots()
}
