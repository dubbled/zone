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
	loc   Point
	owner *Player
	id    string
	x, y  int
	typ   string
	node  *Node
}

func createMember(g Graph, typ string, owner *Player) *Member {
	return &Member{
		id:    uuid.New().String(),
		node:  GetNode(g, owner.x, owner.y),
		owner: owner,
		typ:   typ,
		loc: Point{
			x: owner.x,
			y: owner.y,
		},
	}
}
