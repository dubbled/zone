# Zone

Real-time multiplayer territory control game.

## Tech Stack
- **Backend:** Go 1.24, Chi router, Gorilla WebSocket
- **Frontend:** Vanilla JS, HTML5 Canvas
- **Protocol:** WebSocket for real-time game state sync

## Project Structure
- `main.go` - Server setup, WebSocket handler, game loop
- `player.go` - Player state and resource tracking
- `node.go` - Map tile (node) with resource types
- `member.go` - Game units that players control
- `resources.go` - Resource type definitions
- `static/index.html` - Game client UI

## Development
```bash
go run .
# Open http://localhost:8080
```

## Game Design
- 100x100 tile grid with randomly placed resources (water, wood, metal)
- Players connect via WebSocket and spawn with starting units
- Units move 1 tile per tick toward their target
- Idle units on resource tiles automatically mine resources
- Game ticks once per second; state is broadcast to all connected clients

## Conventions
- Single Go binary, no external DB
- WebSocket is the only client-server communication channel
- Frontend uses vanilla JS with HTML5 Canvas (no frameworks)
- Keep JSON payloads compact (short field names in struct tags)
