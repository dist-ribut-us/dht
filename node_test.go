package dht

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNode(t *testing.T) {
	n := New([]byte{5, 4, 3}, 8)

	ns := []NodeID{
		{128, 100, 123},
		{32, 100, 123},
	}
	n.AddNodeID(ns[0], false)
	n.AddNodeID(ns[1], false)

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
		}
	}
}
