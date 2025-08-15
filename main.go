package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
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

func (s *State) tick() {
	for _, p := range s.players {
		node := s.graph[p.x][p.y]
		member := createMember("unit", node, p)
		node.AddMember(member)
	}
}

// Helper to check if a node has a member
func nodeHasMember(n *Node) bool {
	return len(n.Members) > 0
}

// Find a random point at least minDist away from all players
func findSafeSpawn(graph Graph, minDist int) (int, int) {
	sizeX := len(graph)
	sizeY := len(graph[0])
	tryCount := 0
	for {
		tryCount++
		x := rand.Intn(sizeX)
		y := rand.Intn(sizeY)

		// BFS to check for any player within minDist
		type point struct{ x, y, d int }
		visited := make([][]bool, sizeX)
		for i := range visited {
			visited[i] = make([]bool, sizeY)
		}
		queue := []point{{x, y, 0}}
		found := false
		for len(queue) > 0 {
			p := queue[0]
			queue = queue[1:]
			if p.x < 0 || p.x >= sizeX || p.y < 0 || p.y >= sizeY || visited[p.x][p.y] || p.d > minDist {
				continue
			}
			visited[p.x][p.y] = true
			if p.d > 0 && nodeHasMember(graph[p.x][p.y]) {
				found = true
				break
			}
			// Add neighbors
			queue = append(queue, point{p.x + 1, p.y, p.d + 1})
			queue = append(queue, point{p.x - 1, p.y, p.d + 1})
			queue = append(queue, point{p.x, p.y + 1, p.d + 1})
			queue = append(queue, point{p.x, p.y - 1, p.d + 1})
		}
		if !found {
			fmt.Println("found safe spawn", x, y)
			return x, y
		}
		// else, try again
	}
}

func socketHandler(state *State) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		playerID := chi.URLParam(r, "playerID")
		log.Println("got request for playerID", playerID)

		fmt.Println(state.players)

		player := state.players[playerID]
		if player == nil {
			fmt.Println("finding safe spawn for player")
			x, y := findSafeSpawn(state.graph, 30)
			player = createPlayer(playerID, x, y)
			state.AddPlayer(player)
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
					if err != nil {
						log.Println(err)
						outErrCount++
						if outErrCount > 5 {
							cancel()
							return
						}
					}
					time.Sleep(time.Second * 1)
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
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	w, h := 100, 100
	graph := buildGraph(w, h)
	state := &State{graph: graph, players: map[string]*Player{}}

	// Initialize a fake seed player at random coordinates
	x := rand.Intn(w)
	y := rand.Intn(h)
	seed := createPlayer("seed", x, y)
	state.AddPlayer(seed)

	go func() {
		// Serve static files from the "./static" directory
		r := chi.NewRouter()

		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}))

		fs := http.FileServer(http.Dir("./static"))
		http.Handle("/", fs)

		r.Get("/updates/{playerID}", socketHandler(state))

		log.Println("Starting server at port 8080")
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Fatal(err)
		}
	}()

	for {
		state.tick()
		time.Sleep(time.Second * 1)
	}
}
