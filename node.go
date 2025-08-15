package main

type NodeAttribute string

const (
	NodeWaterSource NodeAttribute = "node_water_source"
	NodeIronSource  NodeAttribute = "node_iron_source"
)

// Node is a single vertex in the graph
type Node struct {
	Members    []*Member        `json:"m"`
	Attributes []*NodeAttribute `json:"a"`
}

func GetNode(g Graph, x, y int) *Node {
	if g.SizeX() <= x || g.SizeY() <= y {
		return nil
	}
	return g[x][y]
}

func (n *Node) AddMember(m *Member) {
	n.Members = append(n.Members, m)
}

func (n *Node) RemoveMember(m *Member) {
	if m.Count > 0 {
		m.Count--
	} else {
		for i, member := range n.Members {
			if member == m {
				n.Members = append(n.Members[:i], n.Members[i+1:]...)
				return
			}
		}
	}
}
