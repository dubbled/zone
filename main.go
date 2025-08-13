package main

import "fmt"

type Graph [][]*Node

type NodeAttribute string

const (
	NodeWaterSource NodeAttribute = "node_water_source"
	NodeIronSource  NodeAttribute = "node_iron_source"
)

// Node is a single vertex in the graph
type Node struct {
	members    []Member
	attributes []NodeAttribute
}

// Player executes actions against the game state
type Player struct{}

// Member is a member of a node, of which there can be many
type Member struct {
	owner *Player
	typ   string
}

type State struct {
	players []*Player
	graph   Graph
}

func (s *State) AddPlayer(p *Player) {
	s.players = append(s.players, p)
}

// An action is an action that a player takes to modify the state
// Actions can be executed over a number of ticks
// Copmlex actions will be reduced by each turn
type Action struct{}

func createNode() *Node {
	return &Node{
		members:    []Member{},
		attributes: []NodeAttribute{},
	}
}

func createPlayer(x, y int) *Player {
	return &Player{}
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
	for _, x := range g {
		for _, y := range x {
			members := len(y.members)
			fmt.Printf("%d", members)
		}
		fmt.Println()
	}
}

func main() {
	graph := buildGraph(10, 10)

	state := &State{graph: graph, players: []*Player{}}

	state.AddPlayer(createPlayer(0, 0))
	state.AddPlayer(createPlayer(len(graph), len(graph[0])))

	printGraph(graph)
}
