package dht

// Node keeps a list of NodeIDs folllowing the linking rules of a DHT.
type Node struct {
	id        NodeID
	blacklist *blacklist
	tree      *tree
}

// New creates a DHT Node
func New(id []byte, startBuffers int) *Node {
	if startBuffers < 1 {
		startBuffers = 1
	}
	return &Node{
		id:        NodeID(id),
		blacklist: newblacklist(),
		tree:      newTree(NodeID(id), startBuffers),
	}
}

// ID returns a copy of the Node's ID.
func (n *Node) ID() NodeID {
	return n.id.Copy()
}

func (n *Node) blacklisted(idStr string) bool {
	b, _ := n.blacklist.get(idStr)
	return b
}

// AddNodeID will add the id to the list of known ids. If the node is
// blacklisted it will not be added unless overrideBlacklist.
func (n *Node) AddNodeID(id NodeID, overrideBlacklist bool) {
	if id == nil || n.id.Equal(id) {
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

// RemoveNodeID removes a NodeID. If blacklist is true, the NodeID will be added
// to the blacklist.
func (n *Node) RemoveNodeID(id NodeID, blacklist bool) {
	if blacklist {
		n.blacklist.set(id.String(), true)
	}
	n.tree.remove(id)
}

// Seek finds the closest node to the target. If mustBeCloser is true it will
// only return a value that is closer to the target than Node.ID
func (n *Node) Seek(target NodeID, mustBeCloser bool) NodeID {
	best := n.tree.search(target)
	if mustBeCloser && n.id.Xor(target).Compare(best.Xor(target)) == -1 {
		return nil
	}
	return best
}

// SeekN searches for multiple NodeIDs close to the target. The max number of
// NodeIDs returned will be ids. If mustBeCloser is true, all returned values
// will be closer to the target than Node.ID
func (n *Node) SeekN(target NodeID, ids int, mustBeCloser bool) []NodeID {
	var c NodeID
	if mustBeCloser {
		c = n.id.Xor(target)
	}
	return n.tree.searchn(target, ids, c)
}

// KnownIDs returns the number of IDs currently stored.
func (n *Node) KnownIDs() int {
	return n.tree.descendants()
}
