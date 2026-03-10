package main

const StartingResources = 100

// Player tracks a connected player's identity and mined resources.
type Player struct {
	ID        string         `json:"id"`
	Resources map[string]int `json:"resources"`
}

func createPlayer(id string) *Player {
	return &Player{
		ID: id,
		Resources: map[string]int{
			"water": StartingResources,
			"wood":  StartingResources,
			"metal": StartingResources,
		},
	}
}
