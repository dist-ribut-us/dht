package dht

import (
	"sort"
	"sync"
)

type nodeIDlist []NodeID

func newNodeIDlist(c int) nodeIDlist {
	if c == -1 {
		c = 0
	}
	return nodeIDlist(make([]NodeID, 0, c))
}

func (n nodeIDlist) Len() int {
	return len(n)
}

func (n nodeIDlist) Get(idx int) NodeID {
	if idx < 0 || idx >= len(n) {
		return nil
	}
	return n[idx]
}

func (n nodeIDlist) insert(id NodeID, idx int, extend bool) nodeIDlist {
	ln := len(n)
	if extend {
		ln++
	}
	nu := nodeIDlist(make([]NodeID, ln))
	if idx < len(nu) {
		copy(nu[:idx], n[:idx])
	}
	copy(nu[idx+1:], n[idx:])
	nu[idx] = id
	return nu
}

func (n nodeIDlist) Search(id NodeID) int {
	idx := sort.Search(len(n), func(i int) bool {
		return id.Compare(n[i]) == -1
	})
	return idx
}

func (n nodeIDlist) Remove(idx int) nodeIDlist {
	ln := len(n) - 1
	if ln < 1 {
		return nil
	}
	nu := nodeIDlist(make([]NodeID, ln))
	copy(nu, n[:idx])
	if idx <= ln {
		copy(nu[idx:], n[idx+1:])
	}
	return nu
}

type List struct {
	target  NodeID
	maxLen  int
	diffs   nodeIDlist
	nodeIDs nodeIDlist
	sync.RWMutex
}

func NewList(target NodeID, maxLen int) *List {
	return &List{
		target:  target,
		maxLen:  maxLen,
		nodeIDs: newNodeIDlist(maxLen),
		diffs:   newNodeIDlist(maxLen),
	}
}

func (l *List) Get(idx int) NodeID {
	return l.nodeIDs.Get(idx)
}

func (l *List) AddNodeIDs(ids []NodeID) {
	for _, id := range ids {
		l.AddNodeID(id)
	}
}

func (l *List) AddNodeID(n NodeID) bool {
	if len(n) != len(l.target) {
		return false
	}

	d := n.Xor(l.target)
	l.RLock()
	diffs, ids := l.diffs, l.nodeIDs
	l.RUnlock()
	idx := diffs.Search(d)
	if ids.Get(idx - 1).Equal(n) {
		return false
	}

	extend := false
	if l.maxLen == -1 || ids.Len() < l.maxLen {
		extend = true
	} else if idx == ids.Len() {
		return false
	}

	l.Lock()
	l.nodeIDs = ids.insert(n, idx, extend)
	l.diffs = diffs.insert(d, idx, extend)
	l.Unlock()
	return true
}

func (l *List) Search(n NodeID) int {
	return l.diffs.Search(n.Xor(l.target))
}

func (l *List) RemoveNodeID(n NodeID) bool {
	d := n.Xor(l.target)
	l.RLock()
	diffs, ids := l.diffs, l.nodeIDs
	l.RUnlock()
	idx := diffs.Search(d)
	if !ids.Get(idx - 1).Equal(n) {
		return false
	}
	l.Lock()
	l.nodeIDs = ids.Remove(idx)
	l.diffs = diffs.Remove(idx)
	l.Unlock()
	return true
}

// Seek returns the closest node to n in the list
func (l *List) Seek(n NodeID) NodeID {
	l.RLock()
	diffs, ids := l.diffs, l.nodeIDs
	l.RUnlock()
	if len(ids) == 0 {
		return nil
	}
	idx := diffs.Search(l.target.Xor(n))
	println(idx)

	if id := ids.Get(idx - 1); id.Equal(n) {
		return id
	}

	return ids.Get(idx)
}

func (l *List) Target() NodeID {
	return l.target.Copy()
}

func (l *List) Len() int {
	return l.nodeIDs.Len()
}
