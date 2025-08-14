package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

type Graph [][]*Node

type State struct {
	players map[string]*Player
	graph   Graph
}

func (s *State) AddPlayer(p *Player) {
	s.players[p.id] = p

	node := s.graph[p.x][p.y]
	member := createMember("player", node, p)

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

func (s *State) tick() {
	// for _, p := range s.players {

	// }
}

func socketHandler(state *State) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		playerID := chi.URLParam(r, "playerID")
		player := state.players[playerID]
		if player == nil {
			http.Error(w, "Player not found", http.StatusNotFound)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		ctx, cancel := context.WithCancel(r.Context())

		type SocketMessage struct {
			Graph [][]*Node `json:"graph"`
		}

		type SocketError struct {
			Error string `json:"error"`
		}

		// write to the socket
		go func() {
			outErrCount := 0

			for {
				select {
				case <-ctx.Done():
					return
				default:
					msg := SocketMessage{Graph: state.graph}
					out, err := json.Marshal(msg)
					if err != nil {
						conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
					}

					err = conn.WriteMessage(websocket.TextMessage, []byte(out))
					// TODO: threshold fail?
					log.Println(err)
					if err != nil {
						outErrCount++
						if outErrCount > 5 {
							cancel()
							return
						}
					}
					time.Sleep(time.Second)
				}
			}
		}()

		// read from the socket
		for {
			select {
			case <-ctx.Done():
				return

			default:
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Println(err)
					return
				}

				actions := []Action{}
				err = json.Unmarshal(message, &actions)
				if err != nil {
					socketError := SocketError{Error: err.Error()}
					conn.WriteMessage(websocket.TextMessage, []byte(socketError.Error))
					continue
				}

				for _, action := range actions {
					player.QueueAction(action)
				}
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	graph := buildGraph(10, 10)

	state := &State{graph: graph, players: map[string]*Player{}}

	state.AddPlayer(createPlayer(0, 0))
	state.AddPlayer(createPlayer(len(graph)-1, len(graph[0])-1))

	go func() {
		// Serve static files from the "./static" directory
		r := chi.NewRouter()
		fs := http.FileServer(http.Dir("./static"))

		http.Handle("/", fs)
		// r.Get("/updates/{playerID}", socketHandler(state))

		log.Println("Starting server at port 8080")
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Fatal(err)
		}
	}()

	for {
		state.tick()
	}
}
