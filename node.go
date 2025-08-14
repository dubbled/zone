package main

type NodeAttribute string

const (
	NodeWaterSource NodeAttribute = "node_water_source"
	NodeIronSource  NodeAttribute = "node_iron_source"
)

// Node is a single vertex in the graph
type Node struct {
	loc        *Point
	members    []*Member
	attributes []*NodeAttribute
}

func GetNode(g Graph, x, y int) *Node {
	if g.SizeX() <= x || g.SizeY() <= y {
		return nil
	}
	return g[x][y]
}

func (n *Node) AddMember(m *Member) {
	n.members = append(n.members, m)
}

func (n *Node) RemoveMember(m *Member) {
	for i, member := range n.members {
		if member.id == m.id {
			n.members = append(n.members[:i], n.members[i+1:]...)
			break
		}
	}
}

func (n *Node) Members() []*Member {
	return n.members
}

func (n *Node) Attributes() []*NodeAttribute {
	return n.attributes
}
