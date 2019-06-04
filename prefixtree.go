package dht

import (
	"fmt"
	"sync"
)

var _ = fmt.Println

type prefixBranch struct {
	descendants uint
	allowed     uint
	toPrune     uint
	branches    [2]*prefixBranch
	val         NodeID
}

func (p *prefixBranch) insert(id NodeID, depth uint) uint {
	if p.descendants == 0 {
		p.val = id
		p.descendants = 1
		return 0
	}
	var valBit byte = 2
	if p.val != nil {
		if p.val.Equal(id) {
			return 0
		}
		valBit = p.val.Bit(depth)
		p.get(valBit).insert(p.val, depth+1)
		p.val = nil
	}
	bit := id.Bit(depth)
	p.toPrune = p.get(bit).insert(id, depth+1)
	bit ^= 1
	if p.branches[bit] != nil && p.toPrune < p.branches[bit].toPrune {
		p.toPrune = p.branches[bit].toPrune
	}

	p.descendants++
	if p.allowed > 0 && p.descendants > p.allowed {
		if toPrune := p.descendants - p.allowed; toPrune > p.toPrune {
			p.toPrune = toPrune
		}
	}

	return p.toPrune
}

func (p *prefixBranch) get(idx byte) *prefixBranch {
	branch := p.branches[idx]
	if branch == nil {
		branch = &prefixBranch{}
		p.branches[idx] = branch
	}
	return branch
}

func (p *prefixBranch) search(target NodeID, depth uint) NodeID {
	if p.val != nil {
		return p.val
	}
	bit := target.Bit(depth)
	if p.branches[bit] != nil && p.branches[bit].descendants > 0 {
		return p.branches[bit].search(target, depth+1)
	}
	bit ^= 1
	if p.branches[bit] != nil && p.branches[bit].descendants > 0 {
		return p.branches[bit].search(target, depth+1)
	}
	return nil
}

func (p *prefixBranch) searchn(target NodeID, ids []NodeID, closerThan NodeID, depth uint) int {
	if p.val != nil {
		if closerThan == nil || p.val.Xor(target).Compare(closerThan) == -1 {
			ids[0] = p.val
			return 1
		} else {
			return 0
		}
	}

	bit := target.Bit(depth)
	var filled int
	if p.branches[bit] != nil {
		filled = p.branches[bit].searchn(target, ids, closerThan, depth+1)
	}
	if filled == len(ids) {
		return filled
	}
	if p.branches[bit^1] != nil {
		filled += p.branches[bit^1].searchn(target, ids[filled:], closerThan, depth+1)
	}
	return filled
}

func (p *prefixBranch) setAllowed(target NodeID, atDepth, allowed, depth uint) {
	b := p.get(target.Bit(depth))
	if depth == atDepth {
		b.allowed = allowed
		return
	}
	b.setAllowed(target, atDepth, allowed, depth+1)
}

//bool indicates if it's safe to remove this branch after pruning
func (p *prefixBranch) prune(n uint, seenAllowed bool) bool {
	canRemove := p.allowed == 0
	if p.allowed > 0 {
		seenAllowed = true
	}

	if p.val != nil {
		if p.descendants != 1 {
			panic("oh boy")
		}
		if n > 0 {
			// if n != 1 {
			// 	println(n)
			// 	panic("really bad")
			// }
			p.val = nil
			p.descendants = 0
		}
		return canRemove && p.descendants == 0
	}
	if p.allowed > 0 && p.toPrune > n {
		n = p.toPrune
		p.toPrune = 0
	}

	var r uint
	p.descendants = 0
	if p.branches[1] != nil {
		if p.branches[1].descendants < n {
			r = n - p.branches[1].descendants
		}
		if p.branches[1].prune(n-r, seenAllowed) {
			p.branches[1] = nil
		} else {
			p.descendants = p.branches[1].descendants
			canRemove = false
		}
	} else {
		r = n
	}

	// if r > 0 && p.branches[0] == nil {
	// 	panic("something bad")
	// }

	if p.branches[0] != nil {
		if p.branches[0].prune(r, seenAllowed) {
			p.branches[0] = nil
		} else {
			canRemove = false
		}
	}

	if p.branches[0] != nil {
		p.descendants += p.branches[0].descendants
	}

	return canRemove && p.descendants == 0
}

func (p *prefixBranch) removeNode(id NodeID, depth uint) {
	if p.val != nil {
		if p.val.Equal(id) {
			p.val = nil
			p.descendants = 0
		}
	}

	bit := id.Bit(depth)
	if p.branches[bit] == nil {
		return
	}
	p.branches[bit].removeNode(id, depth+1)
	p.descendants = p.branches[bit].descendants
	bit ^= 1
	if p.branches[bit] != nil {
		p.descendants += p.branches[bit].descendants
	}
}

type tree struct {
	root         *prefixBranch
	id           NodeID
	toPrune      uint
	startBuffers int
	sync.RWMutex
}

func newTree(id NodeID, startBuffers int) *tree {
	t := &tree{
		root:         &prefixBranch{},
		id:           id,
		startBuffers: startBuffers,
	}

	allowed := uint(startBuffers)
	ln := uint(len(id))*8 - 1
	var i uint
	for ; allowed > 1 && i < ln; i, allowed = i+1, allowed/2 {
		t.root.setAllowed(id.FlipBit(int(i)), i, allowed, 0)
	}
	for ; i < ln; i++ {
		t.root.setAllowed(id.FlipBit(int(i)), i, 1, 0)
	}

	return t
}

func (t *tree) insert(id NodeID) {
	t.Lock()
	t.toPrune = t.root.insert(id, 0)
	t.Unlock()
}

func (t *tree) search(target NodeID) NodeID {
	t.RLock()
	id := t.root.search(target, 0)
	t.RUnlock()
	return id
}

func (t *tree) searchn(id NodeID, n int, closerThan NodeID) []NodeID {
	t.RLock()
	ids := make([]NodeID, n)
	filled := t.root.searchn(id, ids, closerThan, 0)
	t.RUnlock()
	return ids[:filled]
}

func (t *tree) prune() {
	t.Lock()
	t.root.prune(0, false)
	t.Unlock()
}

func (t *tree) Len() int {
	return int(t.root.descendants) - 1
}

func (t *tree) remove(id NodeID) {
	t.Lock()
	t.root.removeNode(id, 0)
	t.Unlock()
}
