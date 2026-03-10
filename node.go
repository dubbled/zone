package main

// Node is a single tile in the game grid.
type Node struct {
	Members  []*Member    `json:"m"`
	Resource ResourceType `json:"r,omitempty"`
}

func createNode(resource ResourceType) *Node {
	return &Node{
		Members:  []*Member{},
		Resource: resource,
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
