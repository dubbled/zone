package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	mrand "math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
)

const (
	GridWidth       = 1000
	GridHeight      = 1000
	TickRate        = time.Second
	SpawnUnits      = 3
	ResourceDensity = 0.15
	MinSpawnDist    = 30
	VisionRadius    = 10
	HouseSpawnRate  = 100 // ticks between house spawns
	MinUnitsForMine = 5
	MineRateMulti   = 3 // mining rate multiplier on mine tiles
)

// Building costs: map[resource]amount
var HouseCost = map[string]int{"wood": 50, "metal": 25}
var MineCost = map[string]int{"wood": 30, "metal": 75}

// Build times (in construction points; each unit contributes 1 per tick)
const (
	HouseBuildTime = 10
	MineBuildTime  = 20
)

// Graph is a 2D grid of nodes.
type Graph [][]*Node

func (g Graph) SizeX() int { return len(g) }
func (g Graph) SizeY() int { return len(g[0]) }

func genMapHash() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// State holds all game state.
type State struct {
	mu        sync.RWMutex
	players   map[string]*Player
	graph     Graph
	members   map[string]*Member
	conns     map[string]*websocket.Conn
	tickCount int
	mapHash   string
}

// Cluster size ranges per resource type.
var clusterSize = map[ResourceType][2]int{
	ResourceWood:  {8, 25},
	ResourceMetal: {3, 12},
	ResourceWater: {10, 30},
}

func buildGraph(w, h int) Graph {
	graph := make(Graph, w)
	for x := range graph {
		graph[x] = make([]*Node, h)
		for y := range graph[x] {
			graph[x][y] = createNode(ResourceNone)
		}
	}

	totalTiles := w * h
	targetResourceTiles := int(float64(totalTiles) * ResourceDensity)
	placed := 0
	types := []ResourceType{ResourceWater, ResourceWood, ResourceMetal}

	for placed < targetResourceTiles {
		res := types[mrand.Intn(len(types))]
		sizeRange := clusterSize[res]
		size := sizeRange[0] + mrand.Intn(sizeRange[1]-sizeRange[0]+1)

		sx := mrand.Intn(w)
		sy := mrand.Intn(h)

		type pt struct{ x, y int }
		frontier := []pt{{sx, sy}}
		visited := map[pt]bool{{sx, sy}: true}
		count := 0

		for len(frontier) > 0 && count < size {
			idx := mrand.Intn(len(frontier))
			p := frontier[idx]
			frontier[idx] = frontier[len(frontier)-1]
			frontier = frontier[:len(frontier)-1]

			if graph[p.x][p.y].Resource != ResourceNone {
				continue
			}

			graph[p.x][p.y].Resource = res
			graph[p.x][p.y].Supply = supplyForResource(res)
			count++

			for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := p.x+d[0], p.y+d[1]
				np := pt{nx, ny}
				if nx >= 0 && nx < w && ny >= 0 && ny < h && !visited[np] {
					visited[np] = true
					frontier = append(frontier, np)
				}
			}
		}
		placed += count
	}

	return graph
}

func findSafeSpawn(graph Graph, members map[string]*Member, minDist int) (int, int) {
	sizeX := graph.SizeX()
	sizeY := graph.SizeY()
	for {
		x := mrand.Intn(sizeX)
		y := mrand.Intn(sizeY)
		safe := true
		for _, m := range members {
			dx := m.x - x
			dy := m.y - y
			if dx*dx+dy*dy < minDist*minDist {
				safe = false
				break
			}
		}
		if safe {
			return x, y
		}
	}
}

func (s *State) addPlayer(id string) *Player {
	player := createPlayer(id)
	s.players[id] = player

	x, y := findSafeSpawn(s.graph, s.members, MinSpawnDist)

	for i := 0; i < SpawnUnits; i++ {
		sx := clamp(x+mrand.Intn(3)-1, 0, GridWidth-1)
		sy := clamp(y+mrand.Intn(3)-1, 0, GridHeight-1)
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

func canAfford(p *Player, cost map[string]int) bool {
	for res, amount := range cost {
		if p.Resources[res] < amount {
			return false
		}
	}
	return true
}

func deductCost(p *Player, cost map[string]int) {
	for res, amount := range cost {
		p.Resources[res] -= amount
	}
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
	s.tickCount++

	// 1. Move units
	s.moveUnits()

	// 2. Advance construction
	s.advanceConstruction()

	// 3. Resolve combat on contested tiles
	s.resolveCombat()

	// 4. House spawning
	s.spawnFromHouses()

	// 5. Mine resources (boosted by mines)
	s.mineResources()

	s.mu.Unlock()
	s.broadcast()
}

func (s *State) moveUnits() {
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
}

func (s *State) advanceConstruction() {
	for x := range s.graph {
		for y := range s.graph[x] {
			node := s.graph[x][y]
			if !node.IsConstructing {
				continue
			}
			workers := node.CountUnitsByOwner(node.BuildingOwner)
			if workers == 0 {
				continue
			}
			node.BuildProgress += workers
			if node.BuildProgress >= node.BuildTarget {
				node.IsConstructing = false
				node.BuildProgress = 0
				node.BuildTarget = 0
				if node.Building == BuildingHouse {
					node.lastSpawnTick = s.tickCount
				}
			}
		}
	}
}

func (s *State) resolveCombat() {
	type pos struct{ x, y int }
	nodeOwners := make(map[pos]map[string][]*Member)

	for _, m := range s.members {
		if m.Typ != "unit" {
			continue
		}
		p := pos{m.x, m.y}
		if nodeOwners[p] == nil {
			nodeOwners[p] = make(map[string][]*Member)
		}
		nodeOwners[p][m.OwnerID] = append(nodeOwners[p][m.OwnerID], m)
	}

	for p, owners := range nodeOwners {
		if len(owners) < 2 {
			continue
		}

		// Find player with most and second-most units
		var maxOwner string
		var maxCount, secondCount int
		for owner, units := range owners {
			if len(units) > maxCount {
				secondCount = maxCount
				maxOwner = owner
				maxCount = len(units)
			} else if len(units) > secondCount {
				secondCount = len(units)
			}
		}

		// Kill all units from losing players
		for owner, units := range owners {
			if owner == maxOwner {
				continue
			}
			for _, m := range units {
				s.graph[p.x][p.y].RemoveMember(m)
				delete(s.members, m.ID)
			}
		}

		// Winner loses units equal to second-highest count
		killed := 0
		for _, m := range owners[maxOwner] {
			if killed >= secondCount {
				break
			}
			s.graph[p.x][p.y].RemoveMember(m)
			delete(s.members, m.ID)
			killed++
		}
	}
}

func (s *State) spawnFromHouses() {
	for x := range s.graph {
		for y := range s.graph[x] {
			node := s.graph[x][y]
			if node.Building != BuildingHouse || node.IsConstructing {
				continue
			}
			if node.CountUnitsByOwner(node.BuildingOwner) < 2 {
				continue
			}
			if s.tickCount-node.lastSpawnTick < HouseSpawnRate {
				continue
			}
			node.lastSpawnTick = s.tickCount
			m := createMember("unit", node.BuildingOwner, x, y, node)
			s.members[m.ID] = m
		}
	}
}

func (s *State) mineResources() {
	for _, m := range s.members {
		if m.HasTarget || m.Typ != "unit" {
			continue
		}
		node := s.graph[m.x][m.y]
		if node.Resource == ResourceNone {
			continue
		}
		player := s.players[m.OwnerID]
		if player == nil {
			continue
		}

		rate := 1
		if node.Building == BuildingMine && !node.IsConstructing && node.BuildingOwner == m.OwnerID {
			rate = MineRateMulti
		}

		resType := string(node.Resource)
		mined := node.Deplete(rate)
		if mined > 0 {
			player.Resources[resType] += mined
		}
	}
}

// VisibleNode is a single tile sent to a client within their fog of war.
type VisibleNode struct {
	X              int          `json:"x"`
	Y              int          `json:"y"`
	Members        []*Member    `json:"m,omitempty"`
	Resource       ResourceType `json:"r,omitempty"`
	Supply         int          `json:"s,omitempty"`
	Building       BuildingType `json:"b,omitempty"`
	BuildingOwner  string       `json:"bo,omitempty"`
	BuildProgress  int          `json:"bp,omitempty"`
	BuildTarget    int          `json:"bt,omitempty"`
	IsConstructing bool         `json:"bc,omitempty"`
}

func (s *State) visibleTiles(playerID string) []VisibleNode {
	seen := make(map[[2]int]bool)
	for _, m := range s.members {
		if m.OwnerID != playerID {
			continue
		}
		x0 := clamp(m.x-VisionRadius, 0, GridWidth-1)
		x1 := clamp(m.x+VisionRadius, 0, GridWidth-1)
		y0 := clamp(m.y-VisionRadius, 0, GridHeight-1)
		y1 := clamp(m.y+VisionRadius, 0, GridHeight-1)
		for x := x0; x <= x1; x++ {
			for y := y0; y <= y1; y++ {
				dx := x - m.x
				dy := y - m.y
				if dx*dx+dy*dy <= VisionRadius*VisionRadius {
					seen[[2]int{x, y}] = true
				}
			}
		}
	}

	tiles := make([]VisibleNode, 0, len(seen))
	for coord := range seen {
		node := s.graph[coord[0]][coord[1]]
		tiles = append(tiles, VisibleNode{
			X:              coord[0],
			Y:              coord[1],
			Members:        node.Members,
			Resource:       node.Resource,
			Supply:         node.Supply,
			Building:       node.Building,
			BuildingOwner:  node.BuildingOwner,
			BuildProgress:  node.BuildProgress,
			BuildTarget:    node.BuildTarget,
			IsConstructing: node.IsConstructing,
		})
	}
	return tiles
}

func (s *State) broadcast() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for pid, conn := range s.conns {
		player := s.players[pid]
		if player == nil {
			continue
		}
		tiles := s.visibleTiles(pid)
		msg, err := json.Marshal(map[string]interface{}{
			"v": tiles,
			"p": player,
		})
		if err != nil {
			log.Println("broadcast marshal error:", err)
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
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
	WriteBufferSize: 65536,
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

		var spawnX, spawnY int
		for _, m := range state.members {
			if m.OwnerID == playerID {
				spawnX = m.x
				spawnY = m.y
				break
			}
		}

		tiles := state.visibleTiles(playerID)
		state.mu.Unlock()

		initMsg, _ := json.Marshal(map[string]interface{}{
			"type":    "init",
			"spawnX":  spawnX,
			"spawnY":  spawnY,
			"v":       tiles,
			"p":       player,
			"mapHash": state.mapHash,
			"costs": map[string]interface{}{
				"house": map[string]interface{}{"resources": HouseCost, "buildTime": HouseBuildTime},
				"mine":  map[string]interface{}{"resources": MineCost, "buildTime": MineBuildTime},
			},
		})
		conn.WriteMessage(websocket.TextMessage, initMsg)

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

			case "build_house":
				x, y := clamp(action.X, 0, GridWidth-1), clamp(action.Y, 0, GridHeight-1)
				node := state.graph[x][y]
				p := state.players[playerID]
				if p != nil && node.Building == BuildingNone && !node.IsConstructing && node.CountUnitsByOwner(playerID) >= 1 && canAfford(p, HouseCost) {
					deductCost(p, HouseCost)
					node.Building = BuildingHouse
					node.BuildingOwner = playerID
					node.IsConstructing = true
					node.BuildProgress = 0
					node.BuildTarget = HouseBuildTime
				}

			case "build_mine":
				x, y := clamp(action.X, 0, GridWidth-1), clamp(action.Y, 0, GridHeight-1)
				node := state.graph[x][y]
				p := state.players[playerID]
				if p != nil && node.Building == BuildingNone && !node.IsConstructing && node.CountUnitsByOwner(playerID) >= MinUnitsForMine && canAfford(p, MineCost) {
					deductCost(p, MineCost)
					node.Building = BuildingMine
					node.BuildingOwner = playerID
					node.IsConstructing = true
					node.BuildProgress = 0
					node.BuildTarget = MineBuildTime
				}
			}
			state.mu.Unlock()
		}

		state.mu.Lock()
		delete(state.conns, playerID)
		state.mu.Unlock()
		log.Println("player disconnected:", playerID)
	}
}

func main() {
	log.Println("building graph...")
	state := &State{
		players: make(map[string]*Player),
		graph:   buildGraph(GridWidth, GridHeight),
		members: make(map[string]*Member),
		conns:   make(map[string]*websocket.Conn),
		mapHash: genMapHash(),
	}
	log.Println("graph ready")

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

	ticker := time.NewTicker(TickRate)
	for range ticker.C {
		state.tick()
	}
}
