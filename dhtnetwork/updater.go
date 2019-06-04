package dhtnetwork

import (
	"crypto/rand"
	"github.com/dist-ribut-us/dht"
	"sync"
)

type action struct {
	idx int
	dht.NodeID
	target dht.NodeID
}

func (u *Updater) seekRequest(a action) (dht.NodeID, SeekRequest) {
	sr := SeekRequest{
		ID:           make([]byte, DefaultIDLen),
		Target:       a.target,
		From:         u.network.NodeID,
		MustBeCloser: true,
	}
	rand.Read(sr.ID)
	u.waiting[encodeToString(sr.ID)] = a
	return a.NodeID, sr
}

type Updater struct {
	network *Node
	queue   []action
	waiting map[string]action
	queued  map[string]bool
	idx     int
	depth   int
	sync.RWMutex
}

func (n *Node) Update() *Updater {
	u := &Updater{
		network: n,
		waiting: make(map[string]action),
		queued:  make(map[string]bool),
		depth:   10,
		idx:     1,
	}

	u.queueIdx(0)

	return u
}

func (u *Updater) queueIdx(idx int) bool {
	target := u.network.NodeID.FlipBit(idx)
	id := u.network.Node.Seek(target, false)
	if id == nil {
		return false
	}
	k := id.String() + target.String()
	u.RLock()
	qd := u.queued[k]
	u.RUnlock()
	if qd {
		return false
	}
	u.Lock()
	u.queued[k] = true
	u.queue = append(u.queue, action{
		target: target,
		NodeID: id,
		idx:    idx,
	})
	u.Unlock()
	return true
}

func (u *Updater) queueLen() int {
	u.RLock()
	l := len(u.queue)
	u.RUnlock()
	return l
}

func (u *Updater) Next() (bool, dht.NodeID, SeekRequest) {
	var ln int
	links := len(u.network.NodeID) * 8

	ln = u.queueLen()

	// By lazy populating the queue, as responses come back, that can be used in
	// later requests.
	for ; ln == 0; ln = u.queueLen() {
		if u.idx >= links {
			return false, nil, SeekRequest{}
		}
		if u.idx > u.depth {
			u.idx = links - 1
		}
		u.queueIdx(u.idx)
		u.idx++
	}
	id, sr := u.seekRequest(u.queue[ln-1])
	u.Lock()
	u.queue = u.queue[:ln-1]
	u.Unlock()
	return true, id, sr
}

// Handle a SeekResponse by adding the nodes to the network.
func (u *Updater) Handle(r SeekResponse) bool {
	idStr := encodeToString(r.ID)
	u.RLock()
	a, found := u.waiting[idStr]
	u.RUnlock()
	if !found {
		return false
	}
	u.Lock()
	delete(u.waiting, idStr)
	u.Unlock()
	u.network.AddNodeID(a.NodeID, true)
	updated := false
	updated = len(r.Nodes) > 0
	u.Lock()
	for _, id := range r.Nodes {
		k := id.String() + a.target.String()
		if u.queued[k] {
			continue
		}
		u.queued[k] = true
		u.queue = append(u.queue, action{
			target: a.target,
			NodeID: id,
			idx:    a.idx,
		})
	}
	u.Unlock()
	return updated
}

func (u *Updater) HandleNoResponse(requestID []byte) {
	idStr := encodeToString(requestID)
	u.RLock()
	a, found := u.waiting[idStr]
	u.RUnlock()
	if !found {
		return
	}
	u.Lock()
	delete(u.waiting, idStr)
	u.Unlock()
	u.network.RemoveNodeID(a.NodeID, true)
	u.queueIdx(a.idx)
}
