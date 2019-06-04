package dhtnetwork

import (
	"crypto/rand"
	"github.com/dist-ribut-us/dht"
	"sort"
)

var DefaultIDLen = 10

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

func (n *Node) Seek(target dht.NodeID) *Seeker {
	s := &Seeker{
		target:     target,
		network:    n,
		sent:       make(map[string]bool),
		reqID2node: make(map[string]dht.NodeID),
	}

	s.Handle(n.HandleSeek(s.seekRequest(n.NodeID, false)))
	return s
}

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
		sr.From = s.network.Copy()
	}
	s.sent[id.String()] = true
	s.reqID2node[encodeToString(sr.ID)] = id
	return sr
}

var MaxQueueDepth = 20

func (s *Seeker) Next() (bool, dht.NodeID, SeekRequest) {
	if s.done {
		return false, nil, SeekRequest{}
	}
	var id dht.NodeID
	for i := 0; i < MaxQueueDepth && i < len(s.queue); i++ {
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
