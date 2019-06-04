package dhtnetwork

import (
	"github.com/dist-ribut-us/dht"
	"github.com/dist-ribut-us/serial"
)

// SeekRequest can be sent to another node on the network as a step in searching
// for a resource.
type SeekRequest struct {
	// Request ID
	ID []byte
	// Target of the seek
	Target dht.NodeID
	// Node that is seeking (so a response can be sent)
	From         dht.NodeID
	MustBeCloser bool
}

var seekRequestPrefixLengths = []int{2, -1, 2, 0}

func (s *SeekRequest) Marshal() ([]byte, error) {
	mustBeCloser := []byte{0}
	if s.MustBeCloser {
		mustBeCloser[0] = 1
	}
	data := [][]byte{
		s.ID,
		mustBeCloser,
		s.Target,
		s.From,
	}
	return serial.MarshalByteSlices(seekRequestPrefixLengths, data)
}

func (s *SeekRequest) Unmarshal(b []byte) error {
	data, err := serial.UnmarshalByteSlices(seekRequestPrefixLengths, b)
	if err != nil {
		return err
	}
	s.ID = data[0]
	s.Target = data[2]
	s.From = data[3]
	if data[1][0] == 1 {
		s.MustBeCloser = true
	}
	return nil
}

// SeekResponse is returned after a SeekRequest with either the data or nodes
// that are closer to the resource.
type SeekResponse struct {
	ID    []byte
	Nodes []dht.NodeID
}

var seekResponsePrefixLengths = []int{1, 0}
var seekResponsePacker = serial.SlicesPacker{
	Count: 2,
	Size:  1,
}

func (s *SeekResponse) Marshal() ([]byte, error) {
	data := make([][]byte, len(s.Nodes))
	for i, id := range s.Nodes {
		data[i] = id
	}
	nbs, err := seekResponsePacker.Marshal(data)
	if err != nil {
		return nil, err
	}
	data = [][]byte{
		s.ID,
		nbs,
	}
	return serial.MarshalByteSlices(seekResponsePrefixLengths, data)
}

func (s *SeekResponse) Unmarshal(b []byte) error {
	data, err := serial.UnmarshalByteSlices(seekResponsePrefixLengths, b)
	if err != nil {
		return err
	}
	s.ID = data[0]
	nbs, err := seekResponsePacker.Unmarshal(data[1])
	if err != nil {
		return err
	}
	s.Nodes = make([]dht.NodeID, len(nbs))
	for i, id := range nbs {
		s.Nodes[i] = id
	}
	return nil
}

// HandleSeek takes a SeekRequest and returns closer nodes up to length
// ReturnNodes. It will not populate Data, instead it is expected that the layer
// managing the hash data checks for the key and only cclosers this if the data is
// not found.
func (n *Node) HandleSeek(r SeekRequest) SeekResponse {
	if !n.SkipRequestUpdate {
		n.AddNodeID(r.From, true)
	}
	// return n.bruteSeek(r)
	return SeekResponse{
		ID:    r.ID,
		Nodes: n.SeekN(r.Target, n.ReturnNodes, r.MustBeCloser),
	}
}

// SearchRange can be used as an Accept function and will return true when one
// of the nodes in the response falls into the range.
func Search(target dht.NodeID) func(SeekResponse) bool {
	return func(sr SeekResponse) bool {
		for _, id := range sr.Nodes {
			if id.Equal(target) {
				return true
			}
		}
		return false
	}
}
