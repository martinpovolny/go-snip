# Ball Fall — Task Plan

## Phase 1: Core Game Logic ✅

- [x] `grid.go` — Grid struct, FindMatches (H/V runs + 2×2), FindMatchGroups (L→R T→B order), ApplyGravity, FillEmpty, Swap, NewGrid
- [x] `protocol.go` — shared message types (StateMsg, MoveMsg, WelcomeMsg, Pos)
- [x] `game.go` — Game state machine (stateWaiting / stateClearing), sequential group processing, cascade
- [x] Unit tests (`grid_test.go`) — H/V runs, 2×2 blocks, gravity, swap, cascade, overlapping patterns, L→R T→B group ordering (10 tests)
- [x] `docs/game-specification.md`, `docs/conventions.md`

## Phase 2: Networking ✅

- [x] `hub.go` — player slot, observer list, Broadcast, DispatchMove; caches lastState for new joiners
- [x] `server_tcp.go` — TCP `:7777` and Unix `/run/ballfall/ballfall.sock`, newline-delimited JSON
- [x] `server_ws.go` — WebSocket `/ws` + HTTP static server `/`
- [x] `cmd/player/main.go` — test player: scans for match-creating swaps in reading order, falls back to random

## Phase 3: Display ✅

- [x] `web/index.html` — canvas grid, WebSocket, click-and-drag moves, flash animation, cascade counter, invalid-move feedback, observer/player badge, touch support
- [x] `game.go` Draw() — ebiten renderer: colored balls, specular highlight, shadow, clearing flash, score + cascade overlay
- [x] `main.go` — CLI flags (--http, --tcp, --unix, --headless), all servers started as goroutines
- [x] `web.go` — web assets embedded via `//go:embed`

## Phase 4: Deployment ✅ (ready, not yet run)

- [x] `Makefile` — `make build / linux-amd64 / deploy / install / status / logs / clean`
- [x] `ballfall.service` — systemd unit, headless mode, RuntimeDirectory=ballfall, WorkingDirectory
- [ ] **Run `make install` on ora-m** — first-time deploy to server

## Phase 5: End-to-end Debug 🔲

- [ ] Run `go run .` + `go run ./cmd/player` together; verify move→clear→cascade→broadcast loop
- [ ] Verify WebSocket browser UI receives board and can send moves
- [ ] Verify observer role (second connection) sees updates but cannot move
- [ ] Fix any bugs found

## Phase 6: Interactive Desktop Input 🔲

- [ ] Ebiten mouse handler — click to select ball, drag to emit move via Hub (currently keyboard/mouse does nothing in the ebiten window; only TCP/WS clients can play)

## Phase 7: Polish 🔲

- [ ] Game-over detection — detect when no valid swap creates a match; broadcast `"status": "gameover"`
- [ ] Observer count shown in web UI
- [ ] Restart/new-game command via socket
