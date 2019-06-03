package dht

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestXor(t *testing.T) {
	a := NodeID{1, 5}
	b := NodeID{2, 2}

	assert.Equal(t, NodeID{3, 7}, a.Xor(b))
}

func TestLeadingZeros(t *testing.T) {
	tt := map[string]struct {
		NodeID
		expected int
	}{
		"Basic": {
			NodeID:   NodeID{128, 0},
			expected: 0,
		},
		"2-Zeros": {
			NodeID:   NodeID{0, 0},
			expected: 16,
		},
		"1,0": {
			NodeID:   NodeID{1, 0},
			expected: 7,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.LeadingZeros())
		})
	}
}

func TestFlipBit(t *testing.T) {
	tt := map[string]struct {
		NodeID
		idx      int
		expected NodeID
	}{
		"{0,0}(0)": {
			NodeID:   NodeID{0, 0},
			idx:      0,
			expected: NodeID{128, 0},
		},
		"{0,0}(7)": {
			NodeID:   NodeID{0, 0},
			idx:      7,
			expected: NodeID{1, 0},
		},
		"{0,0}(8)": {
			NodeID:   NodeID{0, 0},
			idx:      8,
			expected: NodeID{0, 128},
		},
		"{128,0}(0)": {
			NodeID:   NodeID{128, 0},
			idx:      0,
			expected: NodeID{0, 0},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.FlipBit(tc.idx))
		})
	}
}

func TestAdd(t *testing.T) {
	tt := []struct {
		a, b, add, sub NodeID
	}{
		{
			a:   NodeID{1},
			b:   NodeID{2},
			add: NodeID{3},
		},
		{
			a:   NodeID{0, 255},
			b:   NodeID{0, 1},
			add: NodeID{1, 0},
		},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%d %d", tc.a, tc.b), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.add, tc.a.Add(tc.b))
		})
	}
}
