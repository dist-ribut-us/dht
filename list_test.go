package dht

import (
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"testing"
)

const FuzzLoops = 100

func TestSearchDiff(t *testing.T) {
	tt := map[string]struct {
		diffs    []NodeID
		search   NodeID
		expected int
	}{
		"Simple-Match": {
			diffs: []NodeID{
				{0},
				{1},
				{2},
				{3},
			},
			search:   NodeID{2},
			expected: 3,
		},
		"Simple-Gap": {
			diffs: []NodeID{
				{0},
				{1},
				{3},
				{4},
			},
			search:   NodeID{2},
			expected: 2,
		},
		"PastEnd": {
			diffs: []NodeID{
				{0},
				{1},
				{3},
				{4},
			},
			search:   NodeID{100},
			expected: 4,
		},
		"BeforeStart": {
			diffs: []NodeID{
				{1},
				{2},
				{3},
				{4},
			},
			search:   NodeID{0},
			expected: 0,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			n := nodeIDlist(tc.diffs)
			assert.Equal(t, tc.expected, n.Search(tc.search))

			l := &List{
				diffs: n,
			}
			assert.Equal(t, tc.expected, l.diffs.Search(tc.search))
		})
	}
}

func TestAddAndRemoveNodeIDs(t *testing.T) {
	l := NewList(NodeID{1, 4, 16}, 3)

	n := []NodeID{
		{4, 55, 55},
		{2, 55, 55},
		{8, 55, 55},
		{16, 55, 55},
	}

	b, _ := l.AddNodeID(n[0])
	assert.True(t, b)
	b, _ = l.AddNodeID(n[0])
	assert.False(t, b)

	b, _ = l.AddNodeID(n[1])
	assert.True(t, b)
	assert.Equal(t, n[1], l.nodeIDs[0])
	assert.Equal(t, n[1].Xor(l.target), l.diffs[0])
	assert.Equal(t, n[0], l.nodeIDs[1])

	b, _ = l.AddNodeID(n[2])
	assert.True(t, b)
	assert.Equal(t, n[2], l.nodeIDs[2])

	b, _ = l.AddNodeID(n[3])
	assert.False(t, b)

	assert.Len(t, l.nodeIDs, 3)
	assert.True(t, l.RemoveNodeID(l.nodeIDs[len(l.nodeIDs)-1]))
	assert.Len(t, l.nodeIDs, 2)
	assert.Len(t, l.diffs, 2)
	assert.True(t, l.RemoveNodeID(l.nodeIDs[0]))
	assert.Len(t, l.nodeIDs, 1)
	assert.Len(t, l.diffs, 1)

	b, _ = l.AddNodeID(n[3])
	assert.True(t, b)
}

func TestSeek(t *testing.T) {
	l := NewList(NodeID{1, 4, 16}, 3)

	ns := []NodeID{
		{4, 55, 55},
		{16, 55, 55},
		{64, 55, 55},
	}
	for _, n := range ns {
		l.AddNodeID(n)
	}

	assert.Equal(t, ns[0], l.Seek(NodeID{1, 4, 16}))
	// assert.Equal(t, ns[0], l.Seek(NodeID{2, 10, 100}))
	// assert.Equal(t, ns[0], l.Seek(NodeID{4, 55, 54}))
	// assert.Equal(t, ns[0], l.Seek(NodeID{4, 55, 55}))
	// assert.Equal(t, ns[1], l.Seek(NodeID{4, 55, 56}))
	// assert.Equal(t, ns[2], l.Seek(NodeID{64, 55, 55}))
	// assert.Equal(t, ns[2], l.Seek(NodeID{67, 55, 56}))
}

func TestAddFuzz(t *testing.T) {
	ln := 10
	maxLen := 8
	for i := 0; i < FuzzLoops; i++ {
		l := NewList(randID(ln), maxLen)
		for j := 0; j < maxLen*2; j++ {
			l.AddNodeID(randID(ln))
			ok := assert.Equal(t, l.nodeIDs[0].Xor(l.target), l.diffs[0])
			for j := 1; j < len(l.nodeIDs); j++ {
				ok = ok &&
					assert.Equal(t, l.nodeIDs[j].Xor(l.target), l.diffs[j]) &&
					assert.Equal(t, -1, l.diffs[j-1].Compare(l.diffs[j]))
				if !ok {
					t.Log("Failed at ", i, j)
					return
				}
			}
		}
	}
}

func randID(ln int) NodeID {
	id := make(NodeID, ln)
	rand.Read(id)
	return id
}

func TestNodeIDlistInsert(t *testing.T) {
	var n nodeIDlist

	ns := []NodeID{{0}, {1}, {2}, {3}, {4}}

	n = n.insert(ns[1], 0, true)
	assert.Len(t, n, 1)
	assert.Equal(t, ns[1], n[0])

	n = n.insert(ns[0], 0, true)
	assert.Len(t, n, 2)
	assert.Equal(t, ns[0], n[0])
	assert.Equal(t, ns[1], n[1])

	n = n.insert(ns[2], 2, true)
	assert.Len(t, n, 3)
	assert.Equal(t, ns[0], n[0])
	assert.Equal(t, ns[1], n[1])
	assert.Equal(t, ns[2], n[2])
}
