package dhtnetwork

import (
	"encoding/base64"
	"github.com/dist-ribut-us/dht"
)

var encodeToString = base64.URLEncoding.EncodeToString

// Node is our local node in a distributed hash table network.
type Node struct {
	*dht.Node
	ReturnNodes       int
	SkipRequestUpdate bool
	IDlen             int
}

// New creates an instance of Network
func New(self []byte, startBuffers int) *Node {
	return &Node{
		Node:        dht.New(self, startBuffers),
		ReturnNodes: 5,
		IDlen:       len(self),
	}
}
