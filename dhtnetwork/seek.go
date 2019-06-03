package dhtnetwork

import (
	"github.com/dist-ribut-us/dht"
	"github.com/dist-ribut-us/serial"
	"sort"
)

// SeekRequest can be sent to another node on the network as a step in searching
// for a resource.
type SeekRequest struct {
	// Request ID
	ID []byte
	// Target of the seek
	Target dht.NodeID
	// Node that is seeking (so a response can be sent)
	From dht.NodeID
}

var seekRequestPrefixLengths = []int{2, 2, 0}

func (s *SeekRequest) Marshal() ([]byte, error) {
	data := [][]byte{
		s.ID,
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
	s.Target = data[1]
	s.From = data[2]
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
	return n.bruteSeek(r)
}

// There's a more efficient way to do this, but this works. When I write the
// efficient handler, I can use this for fuzzing.
func (n *Node) bruteSeek(r SeekRequest) SeekResponse {
	var closer []dht.NodeID
	d := n.Xor(r.Target)
	ln := n.Links()
	for i := 0; i < ln; i++ {
		for _, id := range n.Link(i) {
			if id.Xor(r.Target).Compare(d) == -1 {
				closer = append(closer, id)
			}
		}
	}
	sort.Slice(closer, func(i, j int) bool {
		di := closer[i].Xor(r.Target)
		dj := closer[i].Xor(r.Target)
		return di.Compare(dj) == -1
	})
	if len(closer) > n.ReturnNodes {
		closer = closer[:n.ReturnNodes]
	}
	return SeekResponse{
		ID:    r.ID,
		Nodes: closer,
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
