package main

import "github.com/google/uuid"

// Player executes actions against the game state
type Player struct {
	id   string
	x, y int
}

func createPlayer(x, y int) *Player {
	p := &Player{id: uuid.New().String(), x: x, y: y}

	return p
}

func (p *Player) MoveMember(g Graph, m *Member) {
	// validate travel path/distance
	g[p.x][p.y].RemoveMember(m)
	m.loc = Point{x: p.x, y: p.y}
	g[p.x][p.y].AddMember(m)
}
