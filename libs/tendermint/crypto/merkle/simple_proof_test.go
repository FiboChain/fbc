package merkle

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestSimpleProofValidateBasic(t *testing.T) {
	testCases := []struct {
		testName      string
		malleateProof func(*SimpleProof)
		errStr        string
	}{
		{"Good", func(sp *SimpleProof) {}, ""},
		{"Negative Total", func(sp *SimpleProof) { sp.Total = -1 }, "negative Total"},
		{"Negative Index", func(sp *SimpleProof) { sp.Index = -1 }, "negative Index"},
		{"Invalid LeafHash", func(sp *SimpleProof) { sp.LeafHash = make([]byte, 10) },
			"expected LeafHash size to be 32, got 10"},
		{"Too many Aunts", func(sp *SimpleProof) { sp.Aunts = make([][]byte, MaxAunts+1) },
			"expected no more than 100 aunts, got 101"},
		{"Invalid Aunt", func(sp *SimpleProof) { sp.Aunts[0] = make([]byte, 10) },
			"expected Aunts#0 size to be 32, got 10"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			_, proofs := SimpleProofsFromByteSlices([][]byte{
				[]byte("apple"),
				[]byte("watermelon"),
				[]byte("kiwi"),
			})
			tc.malleateProof(proofs[0])
			err := proofs[0].ValidateBasic()
			if tc.errStr != "" {
				assert.Contains(t, err.Error(), tc.errStr)
			}
		})
	}
}

func TestSimpleProofAmino(t *testing.T) {
	spTestCases := []SimpleProof{
		{},
		{
			Total:    2,
			Index:    1,
			LeafHash: []byte("LeafHash"),
			Aunts:    [][]byte{[]byte("aunt1"), []byte("aunt2")},
		},
		{
			Total:    math.MaxInt,
			Index:    math.MaxInt,
			LeafHash: []byte{},
			Aunts:    [][]byte{},
		},
		{
			Total: math.MinInt,
			Index: math.MinInt,
			Aunts: [][]byte{nil, {}, []byte("uncle")},
		},
	}

	for _, sp := range spTestCases {
		expectData, err := cdc.MarshalBinaryBare(sp)
		require.NoError(t, err)
		var expectValue SimpleProof
		err = cdc.UnmarshalBinaryBare(expectData, &expectValue)
		require.NoError(t, err)
		var actualValue SimpleProof
		err = actualValue.UnmarshalFromAmino(cdc, expectData)
		require.NoError(t, err)
		require.EqualValues(t, expectValue, actualValue)
	}
}
