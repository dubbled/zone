package main

type Point struct {
	x, y int
}

func createNode() *Node {
	return &Node{
		Members:    []*Member{},
		Attributes: []*NodeAttribute{},
	}
}

// Member is a member of a node, of which there can be many
type Member struct {
	owner *Player
	Typ   string `json:"typ"`
	node  *Node
	Count int `json:"count"`
}

func createMember(typ string, node *Node, owner *Player) *Member {
	member := &Member{
		node:  node,
		owner: owner,
		Typ:   typ,
		Count: 0,
	}

	node.Members = append(node.Members, member)
	return member
}

func (m *Member) AddCount(count int) {
	m.Count += count
}
