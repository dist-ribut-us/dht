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

// Search finds the first id in the list that is greater than the target.
func (n nodeIDlist) Search(target NodeID) int {
	idx := sort.Search(len(n), func(i int) bool {
		return target.Compare(n[i]) == -1
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

func (l *List) AddNodeID(n NodeID) (bool, int) {
	if len(n) != len(l.target) {
		return false, -1
	}

	d := n.Xor(l.target)
	l.Lock()
	diffs, ids := l.diffs, l.nodeIDs
	idx := diffs.Search(d)
	if ids.Get(idx - 1).Equal(n) {
		l.Unlock()
		return false, idx - 1
	}

	extend := false
	if l.maxLen == -1 || ids.Len() < l.maxLen {
		extend = true
	} else if idx == ids.Len() {
		l.Unlock()
		return false, idx
	}

	l.nodeIDs = ids.insert(n, idx, extend)
	l.diffs = diffs.insert(d, idx, extend)
	l.Unlock()
	return true, idx
}

func (l *List) Search(n NodeID) int {
	return l.diffs.Search(n.Xor(l.target))
}

func (l *List) RemoveNodeID(n NodeID) bool {
	d := n.Xor(l.target)
	l.Lock()
	diffs, ids := l.diffs, l.nodeIDs
	idx := diffs.Search(d)
	if !ids.Get(idx - 1).Equal(n) {
		l.Unlock()
		return false
	}
	l.nodeIDs = ids.Remove(idx)
	l.diffs = diffs.Remove(idx)
	l.Unlock()
	return true
}

// Seek returns the closest node to n in the list
func (l *List) Seek(target NodeID) NodeID {
	id, _ := l.bruteSeek(target)
	return id
}

func (l *List) elegantSeek(target NodeID) (NodeID, int) {
	l.RLock()
	diffs, ids := l.diffs, l.nodeIDs
	l.RUnlock()
	ln := len(ids)
	if ln == 0 {
		return nil, -1
	}

	d := target.Xor(l.target)
	idx := diffs.Search(d)
	if idx == ln {
		idx--
	}

	return l.nodeIDs[idx], idx
}

func (l *List) bruteSeek(target NodeID) (NodeID, int) {
	l.RLock()
	ids := l.nodeIDs
	l.RUnlock()
	ln := len(ids)
	if ln == 0 {
		return nil, -1
	}
	if ln == 1 {
		return ids[0], 0
	}

	idx := 0
	best := ids[idx]
	bestd := best.Xor(target)
	for i, id := range ids[1:] {
		d := id.Xor(target)
		if d.Compare(bestd) == -1 {
			best = id
			bestd = d
			idx = i
		}
	}

	return best, idx
}

func (l *List) reorder(target NodeID) []int {
	l2 := NewList(target, l.maxLen)
	out := make([]int, len(l.nodeIDs))
	l2.AddNodeIDs(l.nodeIDs)

	for i := range out {
		out[i] = l2.Search(l.nodeIDs[i]) - 1
	}
	return out
}

func (l *List) Target() NodeID {
	return l.target.Copy()
}

func (l *List) Len() int {
	return l.nodeIDs.Len()
}
