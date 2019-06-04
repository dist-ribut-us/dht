package dht

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTree(t *testing.T) {
	tr := newTree(randID(3), 4)
	ids := []NodeID{
		{10, 20, 30},
		{100, 20, 30},
		{10, 200, 111},
		{100, 200, 111},
	}

	tr.insert(ids[0])
	assert.Equal(t, ids[0], tr.search(ids[0]))

	tr.insert(ids[1])
	assert.Equal(t, ids[0], tr.search(ids[0]))
	assert.Equal(t, ids[1], tr.search(ids[1]))
	assert.Equal(t, ids[0], tr.search(ids[2]))
	assert.Equal(t, ids[1], tr.search(ids[3]))

	s := tr.searchn(ids[0], 5, nil)
	assert.Len(t, s, 2)
	assert.Equal(t, ids[0], s[0])
	assert.Equal(t, ids[1], s[1])

	s = tr.searchn(ids[1], 5, nil)
	assert.Len(t, s, 2)
	assert.Equal(t, ids[1], s[0])
	assert.Equal(t, ids[0], s[1])

	s = tr.searchn(ids[1], 5, NodeID{36, 12, 34})
	assert.Len(t, s, 1)
	assert.Equal(t, ids[1], s[0])

	s = tr.searchn(ids[1], 5, NodeID{110, 0, 1})
	assert.Len(t, s, 2)
	assert.Equal(t, ids[1], s[0])
	assert.Equal(t, ids[0], s[1])

	tr.remove(ids[0])
	s = tr.searchn(ids[1], 5, NodeID{110, 0, 1})
	assert.Len(t, s, 1)
	assert.Equal(t, ids[1], s[0])
}

func TestToPrune(t *testing.T) {
	tr := newTree(NodeID{64, 0, 0}, 32)
	tr.root.setAllowed(NodeID{128 + 64, 0, 0}, 2, 3, 0)

	for i := byte(1); i < 10; i++ {
		tr.insert(NodeID{128 + 64, i, 0})
		if i < 4 {
			assert.EqualValues(t, 0, tr.toPrune)
		} else {
			assert.EqualValues(t, i-3, tr.toPrune)
		}
	}
	tr.insert(NodeID{32, 10, 20})
	assert.EqualValues(t, 6, tr.toPrune)
	tr.insert(NodeID{128, 10, 20})
	assert.EqualValues(t, 6, tr.toPrune)
	tr.insert(NodeID{64, 10, 20})
	assert.EqualValues(t, 6, tr.toPrune)

	tr.root.prune(0, false)
}

func TestFuzzTree(t *testing.T) {
	ln := 10
	for i := 0; i < 1; i++ {
		l := NewList(randID(ln), -1)
		tr := newTree(l.target, 64)

		for j := 0; j < FuzzLoops; j++ {
			id := randID(ln)
			tr.insert(id)
			l.AddNodeID(id)
			assert.Equal(t, j, tr.Len())
		}
		ids := tr.searchn(l.target, FuzzLoops/10, nil)
		assert.Len(t, ids, FuzzLoops/10)
		for j, id := range ids {
			assert.Equal(t, l.nodeIDs[j], id)
		}
		assert.Equal(t, ids, tr.searchn(l.target, FuzzLoops/5, l.diffs[FuzzLoops/10]))

		tr.prune()
		tr.root.checkAllowed()
	}
}

func (p *prefixBranch) checkAllowed() {
	if p.allowed > 0 && p.descendants > p.allowed {
		panic("too many descendants")
	}

	if p.branches[0] != nil {
		p.branches[0].checkAllowed()
	}
	if p.branches[1] != nil {
		p.branches[1].checkAllowed()
	}
}
