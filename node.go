package dht

type Node struct {
	NodeID
	links     []*List
	blacklist *blacklist
	tree      *tree
}

func New(id []byte, startBuffers int) *Node {
	n := &Node{
		NodeID:    NodeID(id),
		blacklist: newblacklist(),
		links:     make([]*List, len(id)*8),
		tree:      newTree(NodeID(id), 64),
	}
	if startBuffers < 1 {
		startBuffers = 1
	}
	i := 0
	for ; startBuffers > 1 && i < len(n.links); i, startBuffers = i+1, startBuffers/2 {
		n.links[i] = NewList(n.FlipBit(i), startBuffers)
	}
	for ; i < len(n.links); i++ {
		n.links[i] = NewList(n.FlipBit(i), 1)
	}
	return n
}

func (n *Node) blacklisted(idStr string) bool {
	b, _ := n.blacklist.get(idStr)
	return b
}

func (n *Node) AddNodeID(id NodeID, overrideBlacklist bool) bool {
	if id == nil || n.Equal(id) {
		return false
	}

	if idStr := id.String(); n.blacklisted(idStr) {
		if overrideBlacklist {
			n.blacklist.delete(idStr)
		} else {
			return false
		}
	}

	n.tree.insert(id)
	if n.tree.toPrune >= uint(n.tree.startBuffers) {
		n.tree.prune()
	}

	d := n.Xor(id)
	lk := n.links[d.LeadingZeros()]
	ins, j := lk.AddNodeID(id)
	needFix := false
	// if there was a race failure, re-insert
	for ins == true {
		x := lk.nodeIDs[j]
		if x.Equal(id) {
			break
		}
		needFix = true
		ins, j = lk.AddNodeID(id)
	}
	if needFix {
		println("fixed race condition")
	}

	return ins
}

func (n *Node) RemoveNodeID(id NodeID, blacklist bool) {
	if blacklist {
		n.blacklist.set(id.String(), true)
	}
	idx := n.Xor(id).LeadingZeros()
	n.links[idx].RemoveNodeID(id)
	n.tree.remove(id)
}

func (n *Node) Seek(target NodeID, mustBeCloser bool) NodeID {
	return n.elegantSeek(target, mustBeCloser)
}

func (n *Node) SeekN(target NodeID, i int, mustBeCloser bool) []NodeID {
	var c NodeID
	if mustBeCloser {
		c = n.Xor(target)
	}
	return n.tree.searchn(target, i, c)
}

func (n *Node) elegantSeek(target NodeID, mustBeCloser bool) NodeID {
	best := n.tree.search(target)
	if mustBeCloser && n.Xor(target).Compare(best.Xor(target)) == -1 {
		return nil
	}
	return best
}

func (n *Node) bruteSeek(target NodeID, mustBeCloser bool) NodeID {
	var best, d NodeID
	ln := n.Links()
	for i := 0; i < ln; i++ {
		for _, id := range n.Link(i) {
			did := id.Xor(target)
			if best == nil || did.Compare(d) == -1 {
				best, d = id, did
			}
		}
	}
	if mustBeCloser && n.Xor(target).Compare(d) == -1 {
		return nil
	}
	return best
}

func (n *Node) Link(idx int) []NodeID {
	return n.links[idx].nodeIDs
}

func (n *Node) Links() int {
	return len(n.links)
}

func (n *Node) LinkTarget(idx int) NodeID {
	return n.links[idx].Target()
}

func (n *Node) KnownIDs() int {
	ln := len(n.links)
	c := 0
	for i := 0; i < ln; i++ {
		c += n.links[i].nodeIDs.Len()
	}
	return c
}
