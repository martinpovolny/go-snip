# Ball Fall — Task Plan

## Phase 1: Core Game Logic ✅

- [x] `grid.go` — Grid struct, FindMatches (H/V runs + 2×2), FindMatchGroups (L→R T→B order), ApplyGravity, FillEmpty, Swap, NewGrid
- [x] `protocol.go` — shared message types (StateMsg, MoveMsg, WelcomeMsg, Pos)
- [x] `game.go` — Game state machine (stateWaiting / stateClearing), sequential group processing, cascade
- [x] Unit tests (`grid_test.go`) — H/V runs, 2×2 blocks, gravity, swap, cascade, overlapping patterns, L→R T→B group ordering (10 tests)
- [x] `docs/game-specification.md`, `docs/conventions.md`

## Phase 2: Networking ✅

- [x] `hub.go` — player slot, observer list, Broadcast, DispatchMove; caches lastState for new joiners
- [x] `server_tcp.go` — TCP `:7777` and Unix `./ballfall.sock`, newline-delimited JSON; claim handling
- [x] `server_ws.go` — WebSocket `/ws` + HTTP static server `/`
- [x] `cmd/connect3/main.go` — test player: scans for match-creating swaps in reading order, falls back to random

## Phase 3: Display ✅

- [x] `web/index.html` — canvas grid, WebSocket, click-and-drag moves, swap/fall animations, cascade counter, invalid-move feedback, observer/player badge, "Take control" claim button
- [x] `draw.go` — ebiten renderer: colored balls, specular highlight, shadow, clearing flash, score + cascade overlay
- [x] `mouse_gui.go` — drag-to-swap mouse handler in ebiten
- [x] `main.go` — CLI flags (--http, --tcp, --unix, --headless), all servers as goroutines

## Phase 4: Deployment ✅

- [x] `Makefile` — `make build / linux-amd64 / deploy / install / status / logs / clean`
- [x] `ballfall.service` — systemd unit, headless mode, RuntimeDirectory=ballfall
- [x] Build tags: default build includes GUI; `-tags nogui` for server (linux) builds

## Phase 5: Integration Tests ✅

- [x] `integration_test.go` — builds nogui binary, starts headless server, tests via TCP and Unix socket
- [x] Tests: Help, TCPConnect, UnixConnect, SecondClientObserver, ObserverClaim, InvalidMove, ValidMove, ValidMoveUnix, Cascade, FallAnimBallData
- [x] Fixed `handleRawConn` writer-goroutine deadlock; added claim handling on TCP; `SendTo` adds newline

## Phase 6: Ball Attack Mode 🔲

### Protocol

- [ ] Add `mode` field (`"demo"` | `"attack"`) to `StateMsg` and `WelcomeMsg`
- [ ] Add `GameOverMsg` (`type: "game_over", score, mode`) to `protocol.go`
- [ ] Add `SetModeMsg` (`type: "set_mode", mode`) to `protocol.go`
- [ ] Handle `set_mode` in `server_tcp.go` and `server_ws.go`; forward to game via new `ModeIn` channel on Hub

### Game Logic

- [ ] Add `gameMode` type (`modeDemo`, `modeAttack`) to `game.go`
- [ ] Add `stateGameOver` to state machine; in this state moves and drops are ignored
- [ ] `NewGrid` variant or `grid.go` helper: `NewAttackGrid()` — fills bottom 2 rows randomly, leaves rows 0–7 empty
- [ ] Ball Attack drop loop: goroutine (or ticker in `Update`) fires every 1 s; picks random color + column; applies slide rule (land on isolated ball → slide L or R); checks game-over condition (column full at row 0) before placing; runs normal match-3 clear/fall/cascade after placement
- [ ] `game.Reset(mode)` — resets grid, score, cascade, state to initial for the chosen mode; broadcasts fresh state
- [ ] Broadcast `GameOverMsg` when Ball Attack ends

### Web UI

- [ ] Add mode toggle buttons to header (next to score); highlight active mode
- [ ] Send `set_mode` on button click
- [ ] Handle `game_over` message: freeze board, show overlay with final score and "Play again / Switch mode" prompt
- [ ] Handle incoming `mode` field in `state` messages to keep buttons in sync

### GUI (ebiten)

- [ ] Expand top margin in `draw.go` to accommodate mode toggle buttons
- [ ] Draw two small mode buttons ("Demo" | "Ball Attack") in the top margin; highlight active
- [ ] `mouse_gui.go`: detect clicks on mode buttons; send `set_mode` via Hub `ModeIn` channel (or call `game.Reset` directly)
- [ ] Ball Attack game-over overlay: centered box with final score text and "Press Space / click to restart" prompt; Space or click calls `game.Reset(modeAttack)`

### Tests

- [ ] Unit tests for `NewAttackGrid()` — correct dimensions, only bottom 2 rows filled
- [ ] Unit test for slide rule — ball landing on isolated stack slides to the correct neighbor
- [ ] Integration test: connect, set Ball Attack mode, verify `mode:"attack"` in state
- [ ] Integration test: fill a column to trigger game-over, verify `game_over` message received
