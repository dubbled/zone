package main

type Resource int

const (
	None Resource = iota
	Brick
	Lumber
	Wool
	Grain
	Ore
)

func (r Resource) String() string {
	return [...]string{"None", "Brick", "Lumber", "Wool", "Grain", "Ore"}[r]
}

type Tile struct {
	q, r     int // axial coordinates
	resource Resource
	number   int // dice number
}

type Position struct {
	q, r int
}

// type Player struct {
// 	id          int
// 	resources   map[Resource]int
// 	settlements map[Position]bool
// 	cities      map[Position]bool
// 	roads       map[[2]Position]bool
// }

type Game struct {
	tiles         []Tile
	players       []*Player
	currentPlayer int
}

// func NewGame() *Game {
// 	rand.Seed(time.Now().UnixNano())
// 	g := &Game{
// 		tiles: make([]Tile, 0),
// 		players: []*Player{
// 			&Player{id: 1, resources: make(map[Resource]int), settlements: make(map[Position]bool), cities: make(map[Position]bool), roads: make(map[[2]Position]bool)},
// 			&Player{id: 2, resources: make(map[Resource]int), settlements: make(map[Position]bool), cities: make(map[Position]bool), roads: make(map[[2]Position]bool)},
// 		},
// 		currentPlayer: 0,
// 	}
// 	g.initBoard()
// 	return g
// }

// func (g *Game) initBoard() {
// 	// Create hex ring 0 (center) and ring 1 (six around)
// 	resourcePool := []Resource{Brick, Lumber, Wool, Grain, Ore, Brick, Lumber, Wool, Grain, Ore, Brick, Lumber}
// 	numbers := []int{5, 2, 6, 3, 8, 10, 9, 12, 11, 4, 8, 10}
// 	// We'll skip 7 for two-player
// 	// For simplicity, we use only 13 tiles
// 	dirs := []Position{{1, 0}, {1, -1}, {0, -1}, {-1, 0}, {-1, 1}, {0, 1}}
// 	idx := 0
// 	// ring 0
// 	g.tiles = append(g.tiles, Tile{q: 0, r: 0, resource: None, number: 0})
// 	// ring 1
// 	for i := 0; i < 6; i++ {
// 		q := dirs[i].q
// 		r := dirs[i].r
// 		res := resourcePool[idx%len(resourcePool)]
// 		num := numbers[idx%len(numbers)]
// 		g.tiles = append(g.tiles, Tile{q: q, r: r, resource: res, number: num})
// 		idx++
// 	}
// }

// func (g *Game) RollDice() int {
// 	return rand.Intn(6) + 1 + rand.Intn(6) + 1
// }

// func (g *Game) DistributeResources(dice int) {
// 	for _, tile := range g.tiles {
// 		if tile.number != dice || tile.resource == None {
// 			continue
// 		}
// 		for _, p := range g.players {
// 			// if player has settlement at any adjacent hex position
// 			// In Catan, settlements are placed at vertices; we skip adjacency for brevity
// 			if p.settlements[Position{tile.q, tile.r}] {
// 				p.resources[tile.resource] += 1
// 			}
// 			if p.cities[Position{tile.q, tile.r}] {
// 				p.resources[tile.resource] += 2
// 			}
// 		}
// 	}
// }

// func (g *Game) BuildSettlement(p *Player, pos Position) bool {
// 	// check if pos is valid and unoccupied
// 	if p.settlements[pos] || p.cities[pos] {
// 		return false
// 	}
// 	// For brevity, we skip distance rule
// 	p.settlements[pos] = true
// 	return true
// }

// func (g *Game) BuildRoad(p *Player, a, b Position) bool {
// 	// check if positions adjacent
// 	if !adjacent(a, b) {
// 		return false
// 	}
// 	// check if road already exists
// 	if p.roads[[2]Position{a, b}] || p.roads[[2]Position{b, a}] {
// 		return false
// 	}
// 	p.roads[[2]Position{a, b}] = true
// 	return true
// }

// func adjacent(a, b Position) bool {
// 	dirs := []Position{{1, 0}, {1, -1}, {0, -1}, {-1, 0}, {-1, 1}, {0, 1}}
// 	for _, d := range dirs {
// 		if a.q+d.q == b.q && a.r+d.r == b.r {
// 			return true
// 		}
// 	}
// 	return false
// }

// func (g *Game) NextTurn() {
// 	g.currentPlayer = (g.currentPlayer + 1) % len(g.players)
// }

// func init() {
// 	game := NewGame()
// 	fmt.Println("Starting 2-player Settlers-like game")
// 	// Place initial settlements for each player
// 	game.BuildSettlement(game.players[0], Position{0, 0})
// 	game.BuildSettlement(game.players[1], Position{1, -1})
// 	// Simple loop of 10 turns
// 	for turn := 0; turn < 10; turn++ {
// 		fmt.Printf("\nTurn %d, Player %d's turn\n", turn+1, game.players[game.currentPlayer].id)
// 		dice := game.RollDice()
// 		fmt.Printf("Rolled %d\n", dice)
// 		game.DistributeResources(dice)
// 		// Show resources
// 		for _, p := range game.players {
// 			fmt.Printf("Player %d resources: %v\n", p.id, p.resources)
// 		}
// 		// Example actions: player builds road if has enough brick and lumber
// 		p := game.players[game.currentPlayer]
// 		if p.resources[Brick] >= 1 && p.resources[Lumber] >= 1 {
// 			p.resources[Brick]--
// 			p.resources[Lumber]--
// 			// Build road from first settlement to adjacent hex
// 			if len(p.settlements) > 0 {
// 				for pos := range p.settlements {
// 					for _, d := range []Position{{1, 0}, {1, -1}, {0, -1}, {-1, 0}, {-1, 1}, {0, 1}} {
// 						newPos := Position{pos.q + d.q, pos.r + d.r}
// 						if game.BuildRoad(p, pos, newPos) {
// 							fmt.Printf("Player %d built a road from %v to %v\n", p.id, pos, newPos)
// 							goto NextTurn
// 						}
// 					}
// 				}
// 			}
// 		}
// 	NextTurn:
// 		game.NextTurn()
// 	}
// 	fmt.Println("\nGame over.")
// }
