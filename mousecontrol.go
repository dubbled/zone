package main

import (
	"fmt"
)

type Resource struct {
	Name string `json:"name"`
	Amount int    `json:"amount"`
}

type State struct {
	players []*Player
	graph   Graph
}

func (s *State) AddPlayer(p *Player, startPos Point) {
	s.players = append(s.players, p)
	p.Draggable = &Draggable{
		Position:  startPos,
		StartPos:  startPos,
		IsDragging: false,
	}
}

type Player struct {
	Name      string
	Draggable *Draggable
}

func (d *Draggable) Drag(x, y float64) {
	d.Position.X = x
	d.Position.Y = y
}

type Point struct {
    X float64
    Y float64
}

func (s *State) OnMouseDown(p *Player, x, y float64) {
	if p.Draggable != nil {
		p.Draggable.IsDragging = true
		p.Draggable.StartPos.X = x
		p.Draggable.StartPos.Y = y
	}
}

func (s *State) OnMouseUp(p *Player) {
	if p.Draggable != nil {
		p.Draggable.IsDragging = false
	}
}

func (s *State) OnMouseMove(p *Player, x, y float64) {
	if p.Draggable != nil && p.Draggable.IsDragging {
		p.Draggable.Drag(x, y)
	}
}

type Draggable struct {
    Position   Point
    StartPos   Point
    IsDragging bool
}

func (s *State) AddPlayer(p *Player) {
    s.players = append(s.players, p)
}

// An action is an action that a player takes to modify the state
// Actions can be executed over a number of ticks
// Complex actions will be reduced by each turn
type Action{}

func createNode() *Node {
    return &Node{
        members:    []Member{},
        attributes: []NodeAttribute{}}}
func createPlayer(id int, name string) *Player {
    return &Player{id: id, name: name}
}

func validateAction(a Action) error {
    return nil
}

func buildGraph(w, h int) Graph {
    graph := make(Graph, w)
    for row, _ := range graph {
        graph[row] = make([]*Node, h)
        column := graph[row]
        for index, _ := range column {
            column[index] = createNode()
        }
    }
    return graph
}

func (g Graph) SizeX() int {
    return len(g)
}

func (g Graph) SizeY() int {
    return len(g[0])
}

func printGraph(g Graph) {
    fmt.Println("Board:")
    for i, row := range g {
        for _, node := range row {
            if len(node.members) > 0 {
                fmt.Printf("%d", node.members[0].owner.id)
            } else {
                fmt.Print(".")
            }
        }
        fmt.Printf("  %d\n", i)
    }
}

func (s *State) MoveMember(fromX, fromY, toX, toY int) error {
    if fromX < 0 || fromY < 0 || fromX >= len(s.graph) || fromY >= len(s.graph[0]) ||
        toX < 0 || toY < 0 || toX >= len(s.graph) || toY >= len(s.graph[0]) {
        return fmt.Errorf("Invalid coordinates")
    }

    fromNode := s.graph[fromX][fromY]
    toNode := s.graph[toX][toY]

    if len(fromNode.members) == 0 {
        return fmt.Errorf("No member to move from the source node")
    }

    memberToMove := fromNode.members[0]
    fromNode.members = fromNode.members[1:]

    toNode.members = append(toNode.members, memberToMove)

    return nil
}

func main() {
    graph := buildGraph(5, 5)
    state := &State{graph: graph, players: []*Player{}}

    // Example usage of the State struct and methods
    player := createPlayer(1, "Alice")
    state.AddPlayer(player, Point{X: 100.0, Y: 200.0})

    fmt.Println("Initial Player Position:", player.Draggable.Position)

    state.OnMouseDown(player, 150.0, 250.0)
    fmt.Println("After Mouse Down:", player.Draggable.Position)

    state.OnMouseMove(player, 200.0, 300.0)
    fmt.Println("After Mouse Move:", player.Draggable.Position)

    state.OnMouseUp(player)
    fmt.Println("After Mouse Up:", player.Draggable.Position)
}
