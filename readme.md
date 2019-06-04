## DHT

This package holds the base logic for a DHT node and the dhtnetwork package
holds the logic for a node to interact with the rest of the network.

There is also a simulation to exercise the whole construction without actually
dealing with networking (IP) logic.

The most important metric is how often Seek can successfully find the resource
it's looking for. Currently, this stands at around 90%.

The number of links fall off exponentionally. But the size of the first set
needs to be proportional to the network. Eventually, this will need to be more
dynamic. A node can estimate the size of the network from the density of node
IDs. That should be used to scale up or down the number of links.

## Prefix Tree
I don't think ordered lists can be used to do an efficient search. But I a
prefix tree can. I've got the insert and searching logic in place.

The next piece is to limit it's growth.

When removing a node, we should consolidate the tree where possible.

## Change search from > to >=
By changing
(n nodeIDlist) Search(target NodeID)
to compare != 1
it would be >= which would remove a lot of the idx-1 checks.
