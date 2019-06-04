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

func (n NodeID) LeadingZeros() int {
	var l int
	for i := 0; i < len(n); i, l = i+1, l+8 {
		if n[i] != 0 {
			b := n[i]
			for ; b < 128; l++ {
				b <<= 1
			}
			break
		}
	}
	return l
}

func (n NodeID) FlipBit(idx int) NodeID {
	out := make(NodeID, len(n))
	copy(out, n)
	out[idx/8] ^= (128 >> uint(idx%8))
	return out
}

func (n NodeID) Bit(idx uint) byte {
	return (n[idx>>3] >> (7 - idx&7)) & 1
}

func (n NodeID) Compare(n2 NodeID) int {
	return bytes.Compare(n, n2)
}

func (n NodeID) Equal(n2 NodeID) bool {
	return bytes.Equal(n, n2)
}

func (n NodeID) String() string {
	return encodeToString(n)
}

func (n NodeID) BitStr() string {
	out := make([]byte, len(n)*9-1)
	for j, b := range n {
		if j != 0 {
			out[j*9-1] = '.'
		}
		for i := 0; i < 8; i++ {
			if b > 127 {
				out[j*9+i] = '1'
			} else {
				out[j*9+i] = '0'
			}
			b <<= 1
		}
	}
	return string(out)
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

// Inc returns the next largest node. Since Seek is exclusive, this can be used
// to do an inclusive seek.
func (n NodeID) Inc() NodeID {
	ln := len(n)
	inc := make(NodeID, ln)
	inc[ln-1] = 1
	return n.Add(inc)
}

func (n NodeID) Copy() NodeID {
	cp := make(NodeID, len(n))
	copy(cp, n)
	return cp
}
