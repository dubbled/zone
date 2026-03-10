package main

// Player tracks a connected player's identity and mined resources.
type Player struct {
	ID        string         `json:"id"`
	Resources map[string]int `json:"resources"`
}

func createPlayer(id string) *Player {
	return &Player{
		ID: id,
		Resources: map[string]int{
			"water": 0,
			"wood":  0,
			"metal": 0,
		},
	}
}
