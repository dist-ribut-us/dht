package dhtnetwork

import (
	"github.com/dist-ribut-us/dht"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandleSeek(t *testing.T) {
	n := New([]byte{1, 10, 15}, 4)

	ns := []dht.NodeID{
		{128, 111, 222},
		{64, 111, 222},
	}
	for _, id := range ns {
		n.AddNodeID(id, false)
	}

	req := SeekRequest{
		ID:     []byte{1, 2, 3},
		Target: dht.NodeID{128 + 64, 111, 222},
	}
	resp := n.HandleSeek(req)
	assert.Equal(t, ns[0], resp.Nodes[0])
	assert.Equal(t, ns[1], resp.Nodes[1])

	req = SeekRequest{
		ID:     []byte{1, 2, 3},
		Target: ns[0],
	}
	resp = n.HandleSeek(req)
	assert.Equal(t, ns[0], resp.Nodes[0])

	req = SeekRequest{
		ID:           []byte{1, 2, 3},
		Target:       ns[1],
		MustBeCloser: true,
	}
	resp = n.HandleSeek(req)
	assert.Equal(t, ns[1], resp.Nodes[0])
}

func TestSeekRequestRoundTrip(t *testing.T) {
	req := SeekRequest{
		ID:     []byte{1, 2, 3},
		Target: dht.NodeID{64, 111, 222},
		From:   dht.NodeID{31, 41, 59},
	}

	b, err := req.Marshal()
	assert.NoError(t, err)

	var out SeekRequest
	assert.NoError(t, out.Unmarshal(b))
	assert.Equal(t, req, out)
}

func TestSeekResponseRoundTrip(t *testing.T) {
	resp := SeekResponse{
		ID: []byte{1, 2, 3},
		Nodes: []dht.NodeID{
			{64, 111, 222},
			{128, 111, 222},
			{64, 111, 222},
		},
	}

	b, err := resp.Marshal()
	assert.NoError(t, err)

	var out SeekResponse
	assert.NoError(t, out.Unmarshal(b))
	assert.Equal(t, resp, out)
}
