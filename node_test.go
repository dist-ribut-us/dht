package dht

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNode(t *testing.T) {
	n := New([]byte{5, 4, 3}, 8)
	assert.Len(t, n.links, 8*3)
	assert.Equal(t, 8, n.links[0].maxLen)
	assert.Equal(t, 4, n.links[1].maxLen)
	assert.Equal(t, 2, n.links[2].maxLen)
	assert.Equal(t, 1, n.links[3].maxLen)
	assert.Equal(t, 1, n.links[4].maxLen)

	ns := []NodeID{
		{128, 100, 123},
		{32, 100, 123},
	}
	n.AddNodeID(ns[0], false)
	assert.Equal(t, ns[0], n.links[0].nodeIDs[0])

	n.AddNodeID(ns[1], false)
	assert.Equal(t, ns[1], n.links[2].nodeIDs[0])

	target := NodeID{48, 213, 222}
	assert.Equal(t, ns[1], n.Seek(target, true))
	assert.Equal(t, ns[0], n.Seek(NodeID{129, 0, 0}, true))
	assert.Nil(t, n.Seek(NodeID{5, 100, 100}, true))
	assert.Nil(t, n.Seek(NodeID{16, 100, 100}, true))
	assert.Equal(t, ns[1], n.Seek(NodeID{32, 213, 222}, true))
}

func TestFuzzNode(t *testing.T) {
	ln := 10
	addNodes := ln * ln * ln
	for i := 0; i < FuzzLoops; i++ {
		n := New(randID(ln), 8)

		for j := 0; j < addNodes; j++ {
			n.AddNodeID(randID(ln), false)
			for lkIdx, lk := range n.links {
				for _, id := range lk.nodeIDs {
					if !assert.Equal(t, lkIdx, n.Xor(id).LeadingZeros()) {
						panic("bad zeros")
					}
				}
			}
		}
	}
}
