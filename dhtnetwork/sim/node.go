package sim

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/dist-ribut-us/dht"
	"github.com/dist-ribut-us/dht/dhtnetwork"
	mr "math/rand"
	"sync"
	"time"
)

var encodeToString = base64.URLEncoding.EncodeToString

const idLen = 32

type seekResponseHandler interface {
	Handle(dhtnetwork.SeekResponse) bool
}

type Node struct {
	net     *dhtnetwork.Node
	gv      *GodView
	send    chan interface{}
	waiting map[string]seekResponseHandler
	sync.RWMutex
}

type stop struct{}
type runUpdate struct{}

func (gv *GodView) AddNode() {
	id := make([]byte, idLen)
	rand.Read(id)

	n := &Node{
		net:     dhtnetwork.New(id, 16),
		gv:      gv,
		send:    make(chan interface{}, 30),
		waiting: make(map[string]seekResponseHandler),
	}
	gv.Lock()
	gv.nodes[n.net.NodeID.String()] = n
	gv.IDs = append(gv.IDs, n.net.NodeID)
	gv.Unlock()

	go n.run()
}

func (n *Node) run() {
	for i := 0; i < 10; i++ {
		n.net.AddNodeID(n.gv.RandID(), true)
	}
	go n.runUpdate()
STOP:
	for {
		switch msg := (<-n.send).(type) {
		case stop:
			break STOP
		case dhtnetwork.SeekRequest:
			n.handleSeekRequest(msg)
		case dhtnetwork.SeekResponse:
			n.handleSeekResponse(msg)
		case runUpdate:
			go n.runUpdate()
		}
	}
	for _ = range n.send {
	}
}

func (n *Node) handleSeekRequest(req dhtnetwork.SeekRequest) {
	n.gv.Send(req.From, n.net.HandleSeek(req))
}

func (n *Node) handleSeekResponse(resp dhtnetwork.SeekResponse) {
	idStr := encodeToString(resp.ID)
	n.RLock()
	h := n.waiting[idStr]
	n.RUnlock()
	if h == nil {
		return
	}
	n.Lock()
	delete(n.waiting, idStr)
	n.Unlock()
	h.Handle(resp)
}

func (n *Node) runUpdate() {
	go func() {
		jitter := time.Duration(float64(n.gv.UpdateFreq) * (0.5 + mr.Float64()))
		time.Sleep(jitter)
		n.send <- runUpdate{}
	}()

	for len(n.net.Link(0)) == 0 {
		n.net.AddNodeID(n.gv.RandID(), true)
	}

	u := n.net.Update()
	for ok, id, sr := u.Next(); ok; ok, id, sr = u.Next() {
		srIDstr := encodeToString(sr.ID)
		n.Lock()
		n.waiting[srIDstr] = u
		n.Unlock()
		if !n.gv.Send(id, sr) {
			n.Lock()
			delete(n.waiting, srIDstr)
			n.Unlock()
			u.HandleNoResponse(sr.ID)
			continue
		}

		notHandled := true
		for i := 0; notHandled == true && i < 20; time.Sleep(time.Millisecond * 4) {
			i++
			n.RLock()
			_, notHandled = n.waiting[srIDstr]
			n.RUnlock()
		}
		if notHandled {
			n.Lock()
			delete(n.waiting, srIDstr)
			n.Unlock()
			u.HandleNoResponse(sr.ID)
		}
	}
}

func (n *Node) Seek(target dht.NodeID) bool {
	s := n.net.Seek(target)
	found := false
	accept := dhtnetwork.Search(target)
	s.Accept = func(resp dhtnetwork.SeekResponse) bool {
		b := accept(resp)
		found = found || b
		return b
	}

	for ok, id, sr := s.Next(); ok; ok, id, sr = s.Next() {
		srIDstr := encodeToString(sr.ID)
		n.Lock()
		n.waiting[srIDstr] = s
		n.Unlock()
		n.gv.Send(id, sr)
		time.Sleep(time.Millisecond * 2)
	}

	return found
}
