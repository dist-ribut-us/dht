package sim

import (
	"fmt"
	"github.com/dist-ribut-us/dht"
	mr "math/rand"
	"sync"
	"time"
)

func init() {
	mr.Seed(time.Now().UnixNano())
}

type GodView struct {
	nodes      map[string]*Node
	IDs        []dht.NodeID
	UpdateFreq time.Duration
	AddFreq    time.Duration
	RemoveFreq time.Duration
	RemoveOdds float64
	SeekFreq   time.Duration
	sync.RWMutex
	commCnt int
}

func New() *GodView {
	return &GodView{
		nodes:      make(map[string]*Node),
		UpdateFreq: time.Millisecond * 1000,
		AddFreq:    time.Millisecond * 200,
		RemoveFreq: time.Second * 3,
		SeekFreq:   time.Millisecond * 1000,
		RemoveOdds: 0.0005,
	}
}

func (gv *GodView) Run() {
	go func() {
		for {
			time.Sleep(gv.AddFreq)
			gv.AddNode()
		}
	}()

	go func() {
		for {
			time.Sleep(gv.RemoveFreq)
			var kept []dht.NodeID
			var kill []*Node
			keptMap := make(map[string]*Node)
			gv.RLock()
			for _, id := range gv.IDs {
				idStr := id.String()
				n := gv.nodes[idStr]
				if mr.Float64() < gv.RemoveOdds {
					kill = append(kill, n)
				} else {
					kept = append(kept, id)
					keptMap[idStr] = n
				}
			}
			gv.RUnlock()
			gv.Lock()
			gv.IDs = kept
			gv.nodes = keptMap
			gv.Unlock()
			for _, n := range kill {
				go gv.send(n, stop{})
			}
			fmt.Println("Nodes in network: ", len(kept))
			fmt.Println("Comms ", gv.commCnt)
			gv.commCnt = 0
		}
	}()

	go func() {
		done := 0
		success := 0
		for {
			time.Sleep(gv.SeekFreq)
			go func() {
				done++
				gv.RLock()
				n := gv.nodes[gv.RandID().String()]
				gv.RUnlock()
				if n == nil {
					return
				}
				b := n.Seek(gv.RandID())
				if b {
					success++
				}
				fmt.Println("Seek: ", b, success, "/", done)
			}()
		}
	}()
}

func (gv *GodView) RandID() dht.NodeID {
	gv.RLock()
	ln := len(gv.IDs)
	if ln == 0 {
		return nil
	}
	id := gv.IDs[mr.Intn(ln)]
	gv.RUnlock()
	return id
}

func (gv *GodView) AddNodes(n int) {
	for i := 0; i < n; i++ {
		gv.AddNode()
	}
}

func (gv *GodView) Send(id dht.NodeID, msg interface{}) bool {
	gv.RLock()
	n := gv.nodes[id.String()]
	gv.RUnlock()
	if n == nil {
		return false
	}
	return gv.send(n, msg)
}

func (gv *GodView) send(n *Node, msg interface{}) bool {
	gv.commCnt++
	select {
	case n.send <- msg:
		return true
	case <-time.After(time.Millisecond * 5):
		println("timeout")
		return false
	}
}
