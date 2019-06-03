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
	sync.RWMutex
}

func (n *Node) Update() *Updater {
	u := &Updater{
		network: n,
		waiting: make(map[string]action),
		queued:  make(map[string]bool),
		idx:     1,
	}

	u.queueIdx(0)

	return u
}

func (u *Updater) queueIdx(idx int) bool {
	target := u.network.LinkTarget(idx)
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

const sendLastAfter = 5

func (u *Updater) Next() (bool, dht.NodeID, SeekRequest) {
	var ln int
	links := u.network.Links()

	ln = u.queueLen()
	if ln == 0 && u.idx > sendLastAfter && u.idx < links {
		// If we're into the sparsly populated tail section, skip ahea and just do
		// the last bucket. This should fill in a few buckets along the way if
		// possible
		sendLast := true
		for i := 0; i < sendLastAfter; i++ {
			if len(u.network.Link(u.idx-i)) > 0 {
				sendLast = false
				break
			}
		}
		// if we just skip to the end always, we should fill some necessary buckets
		// along the way. This nearly works but drops success from ~97% to ~75%.
		if sendLast {
			u.idx = links - 1
		}
	}

	// By lazy populating the queue, as responses come back, that can be used in
	// later requests.
	for ; ln == 0; ln = u.queueLen() {
		if u.idx == links {
			return false, nil, SeekRequest{}
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
	for _, n := range r.Nodes {
		if u.network.AddNodeID(n, false) {
			updated = true
			u.queueIdx(a.idx)
		}
	}
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
