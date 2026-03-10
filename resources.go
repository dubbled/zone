package main

// ResourceType represents a mineable resource on a map tile.
type ResourceType string

const (
	ResourceNone  ResourceType = ""
	ResourceWater ResourceType = "water"
	ResourceWood  ResourceType = "wood"
	ResourceMetal ResourceType = "metal"
)
