package main

import (
	"fmt"
	"log"
	"net/http"
)

type Graph [][]*Node

type State struct {
	players []*Player
	graph   Graph
}

func (s *State) AddPlayer(p *Player) {
	s.players = append(s.players, p)

	member := createMember(s.graph, "player", p)
	s.graph[p.x][p.y].AddMember(member)
}

// An action is an action that a player takes to modify the state
// Actions can be executed over a number of ticks
// Copmlex actions will be reduced by each turn
type Action struct{}

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

	// Serve static files from the "./static" directory
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	go func() {
		log.Println("Starting server at port 8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()

	for {

	}
}
