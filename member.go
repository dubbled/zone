package main

import "github.com/google/uuid"

type Point struct {
	x, y int
}

func createNode() *Node {
	return &Node{
		members:    []*Member{},
		attributes: []*NodeAttribute{},
	}
}

// Member is a member of a node, of which there can be many
type Member struct {
	owner *Player
	id    string
	typ   string
	node  *Node
}

func createMember(typ string, node *Node) *Member {
	member := &Member{
		id:    uuid.New().String(),
		node:  node,
		owner: node.members[0].owner,
		typ:   typ,
	}

	node.members = append(node.members, member)
	return member
}
