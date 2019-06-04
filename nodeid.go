package dht

import (
	"bytes"
	"encoding/base64"
)

// Use URL encoding standard so / doesn't give us trouble
var encodeToString = base64.URLEncoding.EncodeToString
var decodeString = base64.URLEncoding.DecodeString

// NodeID ID represented by a byte slice
type NodeID []byte

// Xor returns a NodeID that is the Xor of the two Nodes
func (n NodeID) Xor(n2 NodeID) NodeID {
	if len(n) != len(n2) {
		return nil
	}
	out := make(NodeID, len(n))
	for i := range out {
		out[i] = n[i] ^ n2[i]
	}
	return out
}

// FlipBit returns a NodeID with the bit at idx flipped
func (n NodeID) FlipBit(idx int) NodeID {
	out := make(NodeID, len(n))
	copy(out, n)
	out[idx>>3] ^= (128 >> uint(idx&7))
	return out
}

// Bit returns the bit at idx
func (n NodeID) Bit(idx uint) byte {
	return (n[idx>>3] >> (7 - idx&7)) & 1
}

// Compare two NodeIds. Returns -1 if n < n2, 0 if n == n2 and 1 if n > n2
func (n NodeID) Compare(n2 NodeID) int {
	return bytes.Compare(n, n2)
}

// Equal reutrns true if n == n2
func (n NodeID) Equal(n2 NodeID) bool {
	return bytes.Equal(n, n2)
}

// String encodes n as a string
func (n NodeID) String() string {
	return encodeToString(n)
}

// Add the node values and handle carry logic from least significant byte (last)
// to most significant (index of 0).
func (n NodeID) Add(n2 NodeID) NodeID {
	if len(n) != len(n2) {
		return nil
	}
	var c uint16
	b := make([]byte, len(n))
	for i := len(n) - 1; i >= 0; i-- {
		v := uint16(n[i]) + uint16(n2[i]) + c
		c = v >> 8
		b[i] = byte(v)
	}
	return b
}

// Copy the NodeID
func (n NodeID) Copy() NodeID {
	cp := make(NodeID, len(n))
	copy(cp, n)
	return cp
}
