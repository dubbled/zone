package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
)

const (
	GridWidth       = 100
	GridHeight      = 100
	TickRate        = time.Second
	SpawnUnits      = 3
	ResourceDensity = 0.15
	MinSpawnDist    = 20
)

// Graph is a 2D grid of nodes.
type Graph [][]*Node

func (g Graph) SizeX() int { return len(g) }
func (g Graph) SizeY() int { return len(g[0]) }

// State holds all game state.
type State struct {
	mu      sync.RWMutex
	players map[string]*Player
	graph   Graph
	members map[string]*Member
	conns   map[string]*websocket.Conn
}

func buildGraph(w, h int) Graph {
	graph := make(Graph, w)
	types := []ResourceType{ResourceWater, ResourceWood, ResourceMetal}
	for x := range graph {
		graph[x] = make([]*Node, h)
		for y := range graph[x] {
			resource := ResourceNone
			if rand.Float64() < ResourceDensity {
				resource = types[rand.Intn(len(types))]
			}
			graph[x][y] = createNode(resource)
		}
	}
	return graph
}

// findSafeSpawn returns coordinates at least minDist tiles from any existing member.
func findSafeSpawn(graph Graph, minDist int) (int, int) {
	sizeX := graph.SizeX()
	sizeY := graph.SizeY()
	for {
		x := rand.Intn(sizeX)
		y := rand.Intn(sizeY)

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
			if p.d > 0 && len(graph[p.x][p.y].Members) > 0 {
				found = true
				break
			}
			queue = append(queue, point{p.x + 1, p.y, p.d + 1})
			queue = append(queue, point{p.x - 1, p.y, p.d + 1})
			queue = append(queue, point{p.x, p.y + 1, p.d + 1})
			queue = append(queue, point{p.x, p.y - 1, p.d + 1})
		}
		if !found {
			return x, y
		}
	}
}

func (s *State) addPlayer(id string) *Player {
	player := createPlayer(id)
	s.players[id] = player

	x, y := findSafeSpawn(s.graph, MinSpawnDist)

	for i := 0; i < SpawnUnits; i++ {
		sx := clamp(x+rand.Intn(3)-1, 0, GridWidth-1)
		sy := clamp(y+rand.Intn(3)-1, 0, GridHeight-1)
		m := createMember("unit", id, sx, sy, s.graph[sx][sy])
		s.members[m.ID] = m
	}

	return player
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func sign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

func (s *State) tick() {
	s.mu.Lock()

	// Collect movements
	type movement struct {
		member       *Member
		fromX, fromY int
		toX, toY     int
	}
	var moves []movement

	for _, m := range s.members {
		if !m.HasTarget {
			continue
		}
		dx := sign(m.TargetX - m.x)
		dy := sign(m.TargetY - m.y)
		newX := clamp(m.x+dx, 0, GridWidth-1)
		newY := clamp(m.y+dy, 0, GridHeight-1)
		if newX != m.x || newY != m.y {
			moves = append(moves, movement{m, m.x, m.y, newX, newY})
		}
	}

	// Apply movements
	for _, mv := range moves {
		s.graph[mv.fromX][mv.fromY].RemoveMember(mv.member)
		mv.member.x = mv.toX
		mv.member.y = mv.toY
		mv.member.node = s.graph[mv.toX][mv.toY]
		s.graph[mv.toX][mv.toY].AddMember(mv.member)

		if mv.member.x == mv.member.TargetX && mv.member.y == mv.member.TargetY {
			mv.member.HasTarget = false
		}
	}

	// Mine resources: idle units on resource tiles earn resources for their owner
	for _, m := range s.members {
		if m.HasTarget || m.Typ != "unit" {
			continue
		}
		node := s.graph[m.x][m.y]
		if node.Resource == ResourceNone {
			continue
		}
		if player := s.players[m.OwnerID]; player != nil {
			player.Resources[string(node.Resource)]++
		}
	}

	s.mu.Unlock()
	s.broadcast()
}

func (s *State) broadcast() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	graphData, err := json.Marshal(s.graph)
	if err != nil {
		log.Println("broadcast marshal error:", err)
		return
	}

	for pid, conn := range s.conns {
		player := s.players[pid]
		if player == nil {
			continue
		}
		playerData, _ := json.Marshal(player)
		msg := fmt.Sprintf(`{"graph":%s,"player":%s}`, string(graphData), string(playerData))
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Println("write error for", pid, err)
		}
	}
}

// ClientAction is a command sent by a player over WebSocket.
type ClientAction struct {
	Type     string `json:"type"`
	MemberID string `json:"member_id"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func socketHandler(state *State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playerID := chi.URLParam(r, "playerID")
		log.Println("player connecting:", playerID)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade error:", err)
			return
		}
		defer conn.Close()

		state.mu.Lock()
		player := state.players[playerID]
		if player == nil {
			player = state.addPlayer(playerID)
		}
		state.conns[playerID] = conn

		// Find spawn position for camera
		var spawnX, spawnY int
		for _, m := range state.members {
			if m.OwnerID == playerID {
				spawnX = m.x
				spawnY = m.y
				break
			}
		}
		state.mu.Unlock()

		// Send init message with spawn position
		state.mu.RLock()
		initMsg, _ := json.Marshal(map[string]interface{}{
			"type":   "init",
			"spawnX": spawnX,
			"spawnY": spawnY,
			"graph":  state.graph,
			"player": player,
		})
		state.mu.RUnlock()
		conn.WriteMessage(websocket.TextMessage, initMsg)

		// Read loop: process player actions
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				break
			}

			var action ClientAction
			if err := json.Unmarshal(message, &action); err != nil {
				log.Println("unmarshal error:", err)
				continue
			}

			state.mu.Lock()
			switch action.Type {
			case "move":
				if member := state.members[action.MemberID]; member != nil && member.OwnerID == playerID {
					member.TargetX = clamp(action.X, 0, GridWidth-1)
					member.TargetY = clamp(action.Y, 0, GridHeight-1)
					member.HasTarget = true
				}
			}
			state.mu.Unlock()
		}

		// Cleanup on disconnect
		state.mu.Lock()
		delete(state.conns, playerID)
		state.mu.Unlock()
		log.Println("player disconnected:", playerID)
	}
}

func main() {
	state := &State{
		players: make(map[string]*Player),
		graph:   buildGraph(GridWidth, GridHeight),
		members: make(map[string]*Member),
		conns:   make(map[string]*websocket.Conn),
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/ws/{playerID}", socketHandler(state))
	r.Handle("/*", http.FileServer(http.Dir("./static")))

	go func() {
		log.Println("server starting on :8080")
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Fatal(err)
		}
	}()

	// Game loop
	ticker := time.NewTicker(TickRate)
	for range ticker.C {
		state.tick()
	}
}
