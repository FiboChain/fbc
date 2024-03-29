// nolint:errcheck
package iavl

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"testing"

	mrand "math/rand"

	cmn "github.com/FiboChain/fbc/libs/iavl/common"
	"github.com/FiboChain/fbc/libs/tendermint/libs/rand"
	db "github.com/FiboChain/fbc/libs/tm-db"
	"github.com/stretchr/testify/require"
	amino "github.com/tendermint/go-amino"
)

type iteratorTestConfig struct {
	startIterate, endIterate     []byte
	startByteToSet, endByteToSet byte
	ascending                    bool
}

func randstr(length int) string {
	return cmn.RandStr(length)
}

func i2b(i int) []byte {
	buf := new(bytes.Buffer)
	amino.EncodeInt32(buf, int32(i))
	return buf.Bytes()
}

func b2i(bz []byte) int {
	i, _, _ := amino.DecodeInt32(bz)
	return int(i)
}

// Construct a MutableTree
func getTestTree(cacheSize int) (*MutableTree, error) {
	return NewMutableTreeWithOpts(db.NewPrefixDB(db.NewMemDB(), []byte(randstr(32))), cacheSize, nil)
}

// Convenience for a new node
func N(l, r interface{}) *Node {
	var left, right *Node
	if _, ok := l.(*Node); ok {
		left = l.(*Node)
	} else {
		left = NewNode(i2b(l.(int)), nil, 0)
	}
	if _, ok := r.(*Node); ok {
		right = r.(*Node)
	} else {
		right = NewNode(i2b(r.(int)), nil, 0)
	}

	n := &Node{
		key:       right.lmd(nil).key,
		value:     nil,
		leftNode:  left,
		rightNode: right,
	}
	n.calcHeightAndSize(nil)
	return n
}

// Setup a deep node
func T(n *Node) *MutableTree {
	t, _ := getTestTree(0)

	n.hashWithCount()
	t.root = n
	return t
}

// Convenience for simple printing of keys & tree structure
func P(n *Node) string {
	if n.height == 0 {
		return fmt.Sprintf("%v", b2i(n.key))
	}
	return fmt.Sprintf("(%v %v)", P(n.leftNode), P(n.rightNode))
}

func randBytes(length int) []byte {
	key := make([]byte, length)
	// math.rand.Read always returns err=nil
	// we do not need cryptographic randomness for this test:
	//nolint:gosec
	mrand.Read(key)
	return key
}

type traverser struct {
	first string
	last  string
	count int
}

func (t *traverser) view(key, value []byte) bool {
	if t.first == "" {
		t.first = string(key)
	}
	t.last = string(key)
	t.count++
	return false
}

func assertMutableMirrorIterate(t *testing.T, tree *MutableTree, mirror map[string]string) {
	sortedMirrorKeys := make([]string, 0, len(mirror))
	for k := range mirror {
		sortedMirrorKeys = append(sortedMirrorKeys, k)
	}
	sort.Strings(sortedMirrorKeys)

	curKeyIdx := 0
	tree.Iterate(func(k, v []byte) bool {
		nextMirrorKey := sortedMirrorKeys[curKeyIdx]
		nextMirrorValue := mirror[nextMirrorKey]

		require.Equal(t, []byte(nextMirrorKey), k)
		require.Equal(t, []byte(nextMirrorValue), v)

		curKeyIdx++
		return false
	})
}

func assertImmutableMirrorIterate(t *testing.T, tree *ImmutableTree, mirror map[string]string) {
	sortedMirrorKeys := getSortedMirrorKeys(mirror)

	curKeyIdx := 0
	tree.Iterate(func(k, v []byte) bool {
		nextMirrorKey := sortedMirrorKeys[curKeyIdx]
		nextMirrorValue := mirror[nextMirrorKey]

		require.Equal(t, []byte(nextMirrorKey), k)
		require.Equal(t, []byte(nextMirrorValue), v)

		curKeyIdx++
		return false
	})
}

func getSortedMirrorKeys(mirror map[string]string) []string {
	sortedMirrorKeys := make([]string, 0, len(mirror))
	for k := range mirror {
		sortedMirrorKeys = append(sortedMirrorKeys, k)
	}
	sort.Strings(sortedMirrorKeys)
	return sortedMirrorKeys
}

func getRandomizedTreeAndMirror(t *testing.T) (*MutableTree, map[string]string) {
	const cacheSize = 100

	tree, err := getTestTree(cacheSize)
	require.NoError(t, err)

	mirror := make(map[string]string)

	randomizeTreeAndMirror(t, tree, mirror)
	return tree, mirror
}

func randomizeTreeAndMirror(t *testing.T, tree *MutableTree, mirror map[string]string) {
	if mirror == nil {
		mirror = make(map[string]string)
	}
	const keyValLength = 5

	numberOfSets := 1000
	numberOfUpdates := numberOfSets / 4
	numberOfRemovals := numberOfSets / 4

	for numberOfSets > numberOfRemovals*3 {
		key := randBytes(keyValLength)
		value := randBytes(keyValLength)

		isUpdated := tree.Set(key, value)
		require.False(t, isUpdated)
		mirror[string(key)] = string(value)

		numberOfSets--
	}

	for numberOfSets+numberOfRemovals+numberOfUpdates > 0 {
		randOp := rand.Intn(3)

		switch randOp {
		case 0:
			if numberOfSets == 0 {
				continue
			}

			numberOfSets--

			key := randBytes(keyValLength)
			value := randBytes(keyValLength)

			isUpdated := tree.Set(key, value)
			require.False(t, isUpdated)
			mirror[string(key)] = string(value)
		case 1:

			if numberOfUpdates == 0 {
				continue
			}
			numberOfUpdates--

			key := getRandomKeyFrom(mirror)
			value := randBytes(keyValLength)

			isUpdated := tree.Set([]byte(key), value)
			require.True(t, isUpdated)
			mirror[key] = string(value)
		case 2:
			if numberOfRemovals == 0 {
				continue
			}
			numberOfRemovals--

			key := getRandomKeyFrom(mirror)

			val, isRemoved := tree.Remove([]byte(key))
			require.True(t, isRemoved)
			require.NotNil(t, val)
			delete(mirror, key)
		default:
			t.Error("Invalid randOp", randOp)
		}
	}
}

func getRandomKeyFrom(mirror map[string]string) string {
	for k := range mirror {
		return k
	}
	panic("no keys in mirror")
}

func setupMirrorForIterator(t *testing.T, config *iteratorTestConfig, tree *MutableTree) [][]string {
	var mirror [][]string

	startByteToSet := config.startByteToSet
	endByteToSet := config.endByteToSet

	if !config.ascending {
		startByteToSet, endByteToSet = endByteToSet, startByteToSet
	}

	curByte := startByteToSet
	for curByte != endByteToSet {
		value := randBytes(5)

		if (config.startIterate == nil || curByte >= config.startIterate[0]) && (config.endIterate == nil || curByte < config.endIterate[0]) {
			mirror = append(mirror, []string{string(curByte), string(value)})
		}

		isUpdated := tree.Set([]byte{curByte}, value)
		require.False(t, isUpdated)

		if config.ascending {
			curByte++
		} else {
			curByte--
		}
	}
	return mirror
}

// assertIterator confirms that the iterator returns the expected values desribed by mirror in the same order.
// mirror is a slice containing slices of the form [key, value]. In other words, key at index 0 and value at index 1.
func assertIterator(t *testing.T, itr db.Iterator, mirror [][]string, ascending bool) {
	startIdx, endIdx := 0, len(mirror)-1
	increment := 1
	mirrorIdx := startIdx

	// flip the iteration order over mirror if descending
	if !ascending {
		startIdx = endIdx - 1
		endIdx = -1
		increment *= -1
	}

	for startIdx != endIdx {
		nextExpectedPair := mirror[mirrorIdx]

		require.True(t, itr.Valid())
		require.Equal(t, []byte(nextExpectedPair[0]), itr.Key())
		require.Equal(t, []byte(nextExpectedPair[1]), itr.Value())
		itr.Next()
		require.NoError(t, itr.Error())

		startIdx += increment
		mirrorIdx++
	}
}

func expectTraverse(t *testing.T, trav traverser, start, end string, count int) {
	if trav.first != start {
		t.Error("Bad start", start, trav.first)
	}
	if trav.last != end {
		t.Error("Bad end", end, trav.last)
	}
	if trav.count != count {
		t.Error("Bad count", count, trav.count)
	}
}

func BenchmarkImmutableAvlTreeMemDB(b *testing.B) {
	db := db.NewDB("test", db.MemDBBackend, "")
	benchmarkImmutableAvlTreeWithDB(b, db)
}

func benchmarkImmutableAvlTreeWithDB(b *testing.B, db db.DB) {
	defer db.Close()

	b.StopTimer()

	t, err := NewMutableTree(db, 100000)
	require.NoError(b, err)

	value := []byte{}
	for i := 0; i < 1000000; i++ {
		t.Set(i2b(int(cmn.RandInt31())), value)
		if i > 990000 && i%1000 == 999 {
			t.SaveVersion(false)
		}
	}
	b.ReportAllocs()
	t.SaveVersion(false)

	runtime.GC()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ri := i2b(int(cmn.RandInt31()))
		t.Set(ri, value)
		t.Remove(ri)
		if i%100 == 99 {
			t.SaveVersion(false)
		}
	}
}
