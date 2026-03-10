package main

// Node is a single tile in the game grid.
type Node struct {
	Members  []*Member    `json:"m"`
	Resource ResourceType `json:"r,omitempty"`
	Supply   int          `json:"s,omitempty"`
}

func createNode(resource ResourceType) *Node {
	return &Node{
		Members:  []*Member{},
		Resource: resource,
		Supply:   supplyForResource(resource),
	}
}

func (n *Node) AddMember(m *Member) {
	n.Members = append(n.Members, m)
}

func (n *Node) RemoveMember(m *Member) {
	for i, member := range n.Members {
		if member.ID == m.ID {
			n.Members = append(n.Members[:i], n.Members[i+1:]...)
			return
		}
	}
}

// Deplete reduces supply by amount and returns actual amount mined.
// When supply hits zero, the resource is cleared.
func (n *Node) Deplete(amount int) int {
	if n.Resource == ResourceNone || n.Supply <= 0 {
		return 0
	}
	mined := amount
	if mined > n.Supply {
		mined = n.Supply
	}
	n.Supply -= mined
	if n.Supply <= 0 {
		n.Resource = ResourceNone
		n.Supply = 0
	}
	return mined
}
