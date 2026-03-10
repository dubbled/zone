package main

// ResourceType represents a mineable resource on a map tile.
type ResourceType string

const (
	ResourceNone  ResourceType = ""
	ResourceWater ResourceType = "water"
	ResourceWood  ResourceType = "wood"
	ResourceMetal ResourceType = "metal"
)

// Starting supply per resource type.
const (
	SupplyWood  = 1_000
	SupplyMetal = 1_000_000
	SupplyWater = 1_000_000
)

func supplyForResource(r ResourceType) int {
	switch r {
	case ResourceWood:
		return SupplyWood
	case ResourceMetal:
		return SupplyMetal
	case ResourceWater:
		return SupplyWater
	default:
		return 0
	}
}
