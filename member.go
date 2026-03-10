package main

import "github.com/google/uuid"

// Member is a unit on the map owned by a player.
type Member struct {
	ID        string `json:"id"`
	OwnerID   string `json:"owner"`
	Typ       string `json:"typ"`
	TargetX   int    `json:"tx"`
	TargetY   int    `json:"ty"`
	HasTarget bool   `json:"moving"`

	// Internal state (not serialized)
	node *Node
	x, y int
}

func createMember(typ string, ownerID string, x, y int, node *Node) *Member {
	m := &Member{
		ID:      uuid.New().String()[:8],
		OwnerID: ownerID,
		Typ:     typ,
		node:    node,
		x:       x,
		y:       y,
	}
	node.AddMember(m)
	return m
}
