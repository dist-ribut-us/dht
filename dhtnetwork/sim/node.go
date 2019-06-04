package sim

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/dist-ribut-us/dht"
	"github.com/dist-ribut-us/dht/dhtnetwork"
	mr "math/rand"
	"time"
)

var encodeToString = base64.URLEncoding.EncodeToString

const idLen = 32

type seekResponseHandler interface {
	Handle(dhtnetwork.SeekResponse) bool
}

type Node struct {
	net           *dhtnetwork.Node
	gv            *GodView
	send          chan interface{}
	waiting       *waiting
	runningUpdate bool
}

type stop struct{}
type runUpdate struct{}

func (gv *GodView) AddNode() {
	id := make([]byte, idLen)
	rand.Read(id)

	n := &Node{
		net:     dhtnetwork.New(id, 64),
		gv:      gv,
		send:    make(chan interface{}, 300),
		waiting: newwaiting(),
	}

	gv.add(n)
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
	h, _ := n.waiting.get(idStr)
	if h == nil {
		return
	}
	n.waiting.delete(idStr)
	h.Handle(resp)
}

func (n *Node) runUpdate() {
	go func() {
		jitter := time.Duration(float64(n.gv.UpdateFreq) * (0.5 + mr.Float64()))
		time.Sleep(jitter)
		n.send <- runUpdate{}
	}()

	if n.runningUpdate {
		return
	}
	n.runningUpdate = true

	for len(n.net.Link(0)) == 0 {
		n.net.AddNodeID(n.gv.RandID(), true)
	}

	u := n.net.Update()
	for ok, id, sr := u.Next(); ok; ok, id, sr = u.Next() {
		srIDstr := encodeToString(sr.ID)
		n.waiting.set(srIDstr, u)
		if !n.gv.Send(id, sr) {
			n.waiting.delete(srIDstr)
			u.HandleNoResponse(sr.ID)
			continue
		}

		notHandled := true
		for i := 0; notHandled == true && i < 20; time.Sleep(time.Millisecond * 4) {
			i++
			_, notHandled = n.waiting.get(srIDstr)
		}
		if notHandled {
			n.waiting.delete(srIDstr)
			u.HandleNoResponse(sr.ID)
		}
	}
	n.runningUpdate = false
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
		n.waiting.set(srIDstr, s)
		n.gv.Send(id, sr)
		time.Sleep(time.Millisecond * 2)

		notHandled := true
		for i := 0; notHandled == true && i < 20; time.Sleep(time.Millisecond * 4) {
			i++
			_, notHandled = n.waiting.get(srIDstr)
		}
		if notHandled {
			n.waiting.delete(srIDstr)
			s.HandleNoResponse(sr.ID)
		}
	}

	return found
}
