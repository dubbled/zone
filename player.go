package main

import (
	"log"

	"github.com/google/uuid"
)

// Player executes actions against the game state
type Player struct {
	id            string
	x, y          int
	QueuedActions []Action
}

func createPlayer(x, y int) *Player {
	p := &Player{id: uuid.New().String(), x: x, y: y}

	return p
}

func (p *Player) validateAction(a Action) error {
	return nil
}

func (p *Player) error(err error) {
	log.Println(err)
}

func (p *Player) QueueAction(a Action) {
	if err := p.validateAction(a); err != nil {
		p.error(err)
		return
	}

	p.QueuedActions = append(p.QueuedActions, a)
}

func (p *Player) MoveMember(g Graph, m *Member, x, y int) {
	// validate travel path/distance
	m.node.RemoveMember(m)
	g[x][y].AddMember(m)
}
