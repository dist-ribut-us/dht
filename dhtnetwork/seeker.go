package dhtnetwork

import (
	"crypto/rand"
	"github.com/dist-ribut-us/dht"
	"sort"
)

// DefaultIDLen is the lenght of SeekRequest IDs
var DefaultIDLen = 10

// Seeker manages Seeking a resource on the network
type Seeker struct {
	target     dht.NodeID
	SkipUpdate bool
	network    *Node
	queue      []dht.NodeID
	sent       map[string]bool
	Accept     func(SeekResponse) bool
	done       bool
	reqID2node map[string]dht.NodeID
	Responses  int
	Successes  int
}

// Seek creates a Seeker for the given target.
func (n *Node) Seek(target dht.NodeID) *Seeker {
	s := &Seeker{
		target:     target,
		network:    n,
		sent:       make(map[string]bool),
		reqID2node: make(map[string]dht.NodeID),
	}

	s.Handle(n.HandleSeek(s.seekRequest(n.ID(), false)))
	return s
}

// Handle a SeekResponse and add the nodes in the response to the queue
func (s *Seeker) Handle(r SeekResponse) bool {
	if s.done == true {
		return false
	}
	rIDstr := encodeToString(r.ID)
	nID, found := s.reqID2node[rIDstr]
	if !found {
		return false
	}
	delete(s.reqID2node, rIDstr)
	if s.network != nil && !s.SkipUpdate {
		s.network.AddNodeID(nID, true)
	}

	for _, id := range r.Nodes {
		s.queue = insert(s.queue, s.target, id)
	}

	if s.Accept != nil && s.Accept(r) {
		s.done = true
	}
	s.Responses++
	s.Successes++
	return true
}

// HandleNoResponse handles the case that a request never got a response.
func (s *Seeker) HandleNoResponse(requestID []byte) {
	if s.done == true {
		return
	}
	_, found := s.reqID2node[encodeToString(requestID)]
	if !found {
		return
	}
	s.Responses++
}

func (s *Seeker) seekRequest(id dht.NodeID, mustBeCloser bool) SeekRequest {
	sr := SeekRequest{
		Target:       s.target,
		ID:           make([]byte, DefaultIDLen),
		MustBeCloser: mustBeCloser,
	}
	rand.Read(sr.ID)
	if s.network != nil {
		sr.From = s.network.ID()
	}
	s.sent[id.String()] = true
	s.reqID2node[encodeToString(sr.ID)] = id
	return sr
}

var maxQueueDepth = 20

// Next returns a bool indication if there this is a valid request, the NodeID
// the request should be sent to and a SeekRequest. It is meant to be used in a
// loop
func (s *Seeker) Next() (bool, dht.NodeID, SeekRequest) {
	if s.done {
		return false, nil, SeekRequest{}
	}
	var id dht.NodeID
	for i := 0; i < maxQueueDepth && i < len(s.queue); i++ {
		if idi := s.queue[i]; !s.sent[idi.String()] {
			id = idi
			break
		}
	}
	if id == nil {
		return false, nil, SeekRequest{}
	}
	return true, id, s.seekRequest(id, true)
}

func insert(q []dht.NodeID, self, id dht.NodeID) []dht.NodeID {
	d := self.Xor(id)
	idx := sort.Search(len(q), func(i int) bool {
		return q[i].Xor(self).Compare(d) != -1
	})
	if idx < len(q) && q[idx].Equal(id) {
		return q
	}
	q = append(q, nil)
	copy(q[idx+1:], q[idx:])
	q[idx] = id
	return q
}
