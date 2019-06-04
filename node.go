package dht

type Node struct {
	NodeID
	blacklist *blacklist
	tree      *tree
}

func New(id []byte, startBuffers int) *Node {
	if startBuffers < 1 {
		startBuffers = 1
	}
	return &Node{
		NodeID:    NodeID(id),
		blacklist: newblacklist(),
		tree:      newTree(NodeID(id), startBuffers),
	}
}

func (n *Node) blacklisted(idStr string) bool {
	b, _ := n.blacklist.get(idStr)
	return b
}

func (n *Node) AddNodeID(id NodeID, overrideBlacklist bool) {
	if id == nil || n.Equal(id) {
		return
	}

	if idStr := id.String(); n.blacklisted(idStr) {
		if overrideBlacklist {
			n.blacklist.delete(idStr)
		} else {
			return
		}
	}

	n.tree.insert(id)
	if n.tree.toPrune >= uint(n.tree.startBuffers) {
		n.tree.prune()
	}
}

func (n *Node) RemoveNodeID(id NodeID, blacklist bool) {
	if blacklist {
		n.blacklist.set(id.String(), true)
	}
	n.tree.remove(id)
}

func (n *Node) Seek(target NodeID, mustBeCloser bool) NodeID {
	best := n.tree.search(target)
	if mustBeCloser && n.Xor(target).Compare(best.Xor(target)) == -1 {
		return nil
	}
	return best
}

func (n *Node) SeekN(target NodeID, i int, mustBeCloser bool) []NodeID {
	var c NodeID
	if mustBeCloser {
		c = n.Xor(target)
	}
	return n.tree.searchn(target, i, c)
}

func (n *Node) KnownIDs() int {
	return n.tree.descendants()
}
