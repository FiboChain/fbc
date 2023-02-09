package mpt

import (
	"github.com/FiboChain/fbc/libs/cosmos-sdk/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	StoreTypeMPT = types.StoreTypeMPT

	TriesInMemory = 100

	// StoreKey is string representation of the store key for mpt
	StoreKey = "mpt"
)

const (
	FlagTrieWriteAhead    = "trie.write-ahead"
	FlagTrieDirtyDisabled = "trie.dirty-disabled"
	FlagTrieCacheSize     = "trie.cache-size"
	FlagTrieNodesLimit    = "trie.nodes-limit"
	FlagTrieImgsLimit     = "trie.imgs-limit"
)

var (
	TrieWriteAhead          = false
	TrieDirtyDisabled       = false
	TrieCacheSize     uint  = 2048 // MB
	TrieNodesLimit    uint  = 256  // MB
	TrieImgsLimit     uint  = 4    // MB
	TrieCommitGap     int64 = 100
)

var (
	KeyPrefixAccRootMptHash        = []byte{0x11}
	KeyPrefixAccLatestStoredHeight = []byte{0x12}
	KeyPrefixEvmRootMptHash        = []byte{0x13}
	KeyPrefixEvmLatestStoredHeight = []byte{0x14}

	GAccToPrefetchChannel    = make(chan [][]byte, 2000)
	GAccTryUpdateTrieChannel = make(chan struct{})
	GAccTrieUpdatedChannel   = make(chan struct{})
)

var (
	NilHash = ethcmn.Hash{}

	// EmptyCodeHash is the known hash of an empty code.
	EmptyCodeHash      = crypto.Keccak256Hash(nil)
	EmptyCodeHashBytes = crypto.Keccak256(nil)

	// EmptyRootHash is the known root hash of an empty trie.
	EmptyRootHash      = ethtypes.EmptyRootHash
	EmptyRootHashBytes = EmptyRootHash.Bytes()
)
