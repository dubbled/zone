package main

// BuildingType represents a player-constructed building on a tile.
type BuildingType string

const (
	BuildingNone  BuildingType = ""
	BuildingHouse BuildingType = "house"
	BuildingMine  BuildingType = "mine"
)

// Node is a single tile in the game grid.
type Node struct {
	Members        []*Member    `json:"m"`
	Resource       ResourceType `json:"r,omitempty"`
	Supply         int          `json:"s,omitempty"`
	Building       BuildingType `json:"b,omitempty"`
	BuildingOwner  string       `json:"bo,omitempty"`
	BuildProgress  int          `json:"bp,omitempty"`
	BuildTarget    int          `json:"bt,omitempty"`
	IsConstructing bool         `json:"bc,omitempty"`

	lastSpawnTick int // internal: tick when house last spawned a unit
}

func createNode(resource ResourceType) *Node {
	return &Node{
		Members:  []*Member{},
		Resource: resource,
		Supply:   supplyForResource(resource),
	}
}

func (n *Node) AddMember(m *Member) {
	n.Members = append(n.Members, m)
}

func (n *Node) RemoveMember(m *Member) {
	for i, member := range n.Members {
		if member.ID == m.ID {
			n.Members = append(n.Members[:i], n.Members[i+1:]...)
			return
		}
	}
}

// Deplete reduces supply by amount and returns actual amount mined.
// When supply hits zero, the resource is cleared.
func (n *Node) Deplete(amount int) int {
	if n.Resource == ResourceNone || n.Supply <= 0 {
		return 0
	}
	mined := amount
	if mined > n.Supply {
		mined = n.Supply
	}
	n.Supply -= mined
	if n.Supply <= 0 {
		n.Resource = ResourceNone
		n.Supply = 0
	}
	return mined
}

// CountUnitsByOwner returns the number of "unit" type members for a given owner.
func (n *Node) CountUnitsByOwner(ownerID string) int {
	count := 0
	for _, m := range n.Members {
		if m.Typ == "unit" && m.OwnerID == ownerID {
			count++
		}
	}
	return count
}
