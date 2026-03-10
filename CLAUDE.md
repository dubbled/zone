# Zone

Real-time multiplayer territory control game.

## Tech Stack
- **Backend:** Go 1.24, Chi router, Gorilla WebSocket
- **Frontend:** Vanilla JS, HTML5 Canvas
- **Protocol:** WebSocket for real-time game state sync

## Project Structure
- `main.go` - Server setup, WebSocket handler, game loop, graph building
- `player.go` - Player state and resource tracking
- `node.go` - Map tile with resource type, supply, and depletion
- `member.go` - Game units that players control
- `resources.go` - Resource type definitions and supply constants
- `static/index.html` - Game client UI

## Development
```bash
go run .
# Open http://localhost:8080
```

## Game Design
- 1000x1000 tile grid with clustered resources (water, wood, metal)
- Players spawn with 3 units and 100 of each resource
- Units move 1 tile per tick toward their target
- Idle units on resource tiles mine until depleted (wood=1K, water/metal=1M)
- Fog of war: 10-tile vision radius per unit
- Map state persisted client-side via localStorage
- Game ticks once per second; only fog-visible tiles sent per player

## Conventions
- Single Go binary, no external DB
- WebSocket is the only client-server communication channel
- Frontend uses vanilla JS with HTML5 Canvas (no frameworks)
- Keep JSON payloads compact (short field names in struct tags)
- Sparse tile broadcast for performance on large grids

## Git Workflow
- Always checkout to verify correct branch before committing and pushing
- Commit and push after completing changes
