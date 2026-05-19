# Ball Fall — Game Specification

## Overview

A match-3 gravity game. A rectangular grid is filled with colored balls. A player swaps adjacent balls; when 3 or more of the same color align (or form a 2×2 block) they disappear, remaining balls fall due to gravity, and new random balls fill the top. Cascade chains score bonus points.

Two game modes share the same match-3 swap mechanics. The player selects a mode before play begins.

## Grid

| Parameter | Value |
|-----------|-------|
| Width     | 8 columns |
| Height    | 10 rows |
| Cell size | 60 px (display) |
| Colors    | Red (1), Green (2), Blue (3), Yellow (4) |
| Empty     | 0 |

Row 0 is the top. New balls appear at row 0 and fall downward.

## Game Modes

### Demo (default)

Endless match-3. The grid starts fully populated (including 4 randomly placed brick obstacles). No game-over condition. Score accumulates until the player resets by switching mode.

### Ball Attack

A survival mode. Balls fall from the top at a fixed rate; the player must clear them by swapping to form matches. The game ends when a column is too full to accept the next incoming ball. No bricks in Ball Attack.

**Starting state:** the bottom 2 rows (rows 8–9) are pre-filled with random balls, guaranteed match-free; rows 0–7 are empty.

**Drop mechanic (every 1 second while in `stateWaiting`):**

1. A ball of random color is chosen.
2. A column is chosen at random.
3. The ball lands at the topmost empty row in that column (the first occupied row minus one).
4. **Game over check:** if the column is already full (`landRow < 0`), a `game_over` message is broadcast and the game freezes.
5. **Slide rule:** if the ball lands directly on top of the stack (a ball exists immediately below) and there are no horizontal neighbors at the landing row, the ball slides one step left or right (direction chosen randomly; tries the other direction if the first destination is also full).
6. After placement the match-3 clearing logic runs exactly as in Demo (find groups → clear → gravity → cascade). The drop timer resets after each drop.

**Note:** the drop timer only advances during `stateWaiting`. It pauses automatically while swap, clear, or fall animations are in progress.

**Game over:** a `game_over` message is broadcast; the board freezes and no further drops or player moves are accepted until the player re-selects a mode.

**Scoring:** same rules as Demo. Score resets on each new session (either mode).

## Bricks

Demo mode places 4 brick obstacles at random positions. Bricks are immovable — balls cannot be swapped into or out of a brick cell, and gravity treats bricks as boundaries (each contiguous column segment between bricks compacts independently).

When a match is cleared, every brick orthogonally adjacent to a matched cell is destroyed. Gravity and match detection run after brick destruction.

Bricks are not present in Ball Attack mode.

## Match Rules

Matches are detected after every swap (and after every cascade):

1. **Horizontal run**: 3 or more consecutive same-color balls in a row.
2. **Vertical run**: 3 or more consecutive same-color balls in a column.
3. **2×2 block**: any 2×2 area where all four cells share the same color.

All matching cells across all rules are collected into one set and cleared simultaneously. Bricks do not participate in matches.

## Game Flow

### Demo

```
stateWaiting
  ↓ receive move
stateSwapping / stateSwapBack   (swap animation, 12 frames)
  → no match found → swap reverted → stateWaiting  (board broadcast)
  → matches found  → stateClearing
stateClearing                   (30-frame flash animation)
  ↓ timer expires → remove matched cells, destroy adjacent bricks
                  → apply gravity (GravityAndFill: compact + refill top)
stateFalling                    (physics fall animation)
  → new matches found → stateClearing  (cascade, g.cascade++)
  → no matches        → stateWaiting   (board broadcast)
```

### Ball Attack

Same swap/clear/fall states as Demo, plus:

```
(every 1 s while in stateWaiting)
  drop ticker fires
  → chosen column full (landRow < 0) → stateGameOver  (broadcast game_over)
  → ball placed (with optional slide)
      → match check → clear/fall/cascade as normal   (GravityOnly: no refill)
```

## Player Moves

A move specifies a cell by `(row, col)` and a direction (`left`, `right`, `up`, `down`). The game swaps the selected cell with its neighbor in that direction. If the resulting board has no matches the swap is immediately reversed (invalid). A `swap_anim` message is broadcast for both valid and invalid swaps; its `valid` field indicates the outcome.

Moves are ignored in `stateGameOver`. Moves may be queued (hub buffer: 4) while a clear or fall animation is in progress; they are applied when `stateWaiting` resumes.

Only one player can be active at a time. The first client to connect becomes the player. If that client disconnects the next observer is automatically promoted. Any observer can also claim the player role explicitly.

## Scoring

| Event | Points |
|-------|--------|
| Each ball cleared | +1 per ball |
| Cascade depth | `cascade` counter; display shows `×(cascade+1)` |

Score resets when a new game session starts (either mode).

## Mode Selection

### Web UI

A pair of buttons in the header bar next to the score. The active mode is highlighted. Clicking the inactive mode sends a `set_mode` message, which resets the game server-side and broadcasts a fresh state.

### GUI (ebiten)

Two small toggle buttons in the top HUD margin, always visible. The active mode is highlighted. Clicking the inactive mode calls `game.Reset(mode)` directly. When Ball Attack reaches game-over a centered overlay shows the final score; clicking anywhere on it (or pressing Space) restarts Ball Attack.

---

## Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                         game binary                              │
│                                                                  │
│  ┌────────────────────┐          ┌───────────────────────────┐   │
│  │  ebiten Display    │          │          Hub              │   │
│  │  (GUI only)        │          │  ┌──────────────────────┐ │   │
│  │  draw.go           │          │  │  clients map         │ │   │
│  │  mouse_gui.go      │          │  │  player *Client      │ │   │
│  └────────┬───────────┘          │  │  lastState []byte    │ │   │
│           │ handleMouse()        │  │  MoveIn chan         │ │   │
│           ▼                      │  │  ModeIn chan         │ │   │
│  ┌────────────────────┐          │  └──────────────────────┘ │   │
│  │  Game / Grid       │─Broadcast│                           │   │
│  │  game.go / grid.go │─────────▶│  ┌──────────────────┐     │   │
│  │  LogicTick() 60 Hz │◀─MoveIn──│  │  server_tcp.go   │     │   │
│  └────────────────────┘◀─ModeIn──│  │  TCP  :7777      │◀──────── cmd/connect3
│                                  │  │  Unix ./ballfall │     │    (test player)
│                                  │  │       .sock      │     │   │
│                                  │  └──────────────────┘     │   │
│                                  │  ┌──────────────────┐     │   │
│                                  │  │  server_ws.go    │     │   │
│                                  │  │  WS   /ws        │◀──────── browser
│                                  │  │  HTTP /          │     │    (web/index.html)
│                                  │  └──────────────────┘     │   │
│                                  └───────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘

Integration tests (integration_test.go, build tag: integration)
  └─ builds headless binary → starts server on ephemeral ports
     → connects via TCP and Unix socket → asserts protocol messages
```

**Key goroutines at runtime:**

| Goroutine | Source | Role |
|-----------|--------|------|
| ebiten main thread | `gui.go` | Calls `handleMouse()` each frame, renders via `draw.go` |
| Logic ticker (60 Hz) | `gui.go` | Calls `game.LogicTick()` on a fixed-rate `time.Ticker` |
| TCP listener | `server_tcp.go` | Accepts TCP connections on `:7777` |
| Unix listener | `server_tcp.go` | Accepts Unix socket connections |
| per-connection reader | `server_tcp.go` | Reads JSON lines from one client |
| per-connection writer | `server_tcp.go` | Drains client's `send` channel |
| WebSocket + HTTP | `server_ws.go` | Handles `/ws` upgrades and static `/` |
| per-WS reader/writer | `server_ws.go` | One goroutine each per WebSocket client |

---

## Communication Protocol (JSON, newline-delimited)

All messages are JSON objects terminated by `\n`. The same protocol is used over TCP, Unix sockets, and WebSocket.

### Server → Client Messages

#### `state` — full board snapshot

Sent when the game enters `stateWaiting` or `stateClearing`, and to new clients on connect (as `lastState` cached by Hub).

```json
{
  "type":    "state",
  "board":   [[1,2,3,4,2,1,3,4], ...],
  "bricks":  [[false,false,true,false,...], ...],
  "width":   8,
  "height":  10,
  "score":   150,
  "status":  "waiting",
  "matches": [[3,4],[3,5],[3,6]],
  "cascade": 0,
  "mode":    "demo"
}
```

| Field | Values | Notes |
|-------|--------|-------|
| `status` | `"waiting"` \| `"clearing"` | |
| `mode` | `"demo"` \| `"attack"` | |
| `matches` | `[[row,col],...]` | Only present when `status == "clearing"` |
| `bricks` | `[[bool,...],...]` | `true` = brick at that cell |

#### `swap_anim` — swap animation

Broadcast immediately when a swap is initiated. Clients play the forward animation; `valid: false` means they should also play the reverse (snap-back).

```json
{
  "type":   "swap_anim",
  "rowA":   3, "colA": 4,
  "rowB":   3, "colB": 5,
  "colorA": 2, "colorB": 1,
  "valid":  true
}
```

#### `fall_anim` — gravity/fall animation

Broadcast when gravity runs after a match clear. `board` and `bricks` are the final state after gravity and refill. Clients animate the listed balls falling, then apply the board snapshot.

```json
{
  "type":    "fall_anim",
  "board":   [[...], ...],
  "bricks":  [[...], ...],
  "balls":   [
    {"color": 2, "col": 3, "fromRow": -1, "toRow": 5},
    {"color": 1, "col": 5, "fromRow": 2,  "toRow": 7}
  ],
  "score":   160,
  "cascade": 1
}
```

`fromRow` may be negative for balls spawned above the grid (Demo mode refill).

#### `welcome` — sent once on connect

```json
{"type": "welcome", "role": "player", "width": 8, "height": 10, "mode": "demo"}
```

#### `role` — role change notification

Sent to an individual client when its role changes (e.g. after being claimed or auto-promoted).

```json
{"type": "role", "role": "observer"}
```

#### `game_over` — Ball Attack session ended

```json
{"type": "game_over", "score": 342, "mode": "attack"}
```

### Client → Server Messages

#### `move` — swap request (player only)

```json
{"type": "move", "row": 3, "col": 4, "dir": "right"}
```

`dir`: `"left"` | `"right"` | `"up"` | `"down"`. Ignored in `stateGameOver`.

#### `set_mode` — switch game mode (player only)

```json
{"type": "set_mode", "mode": "attack"}
```

Resets the game, broadcasts a fresh `state` to all clients.

#### `claim` — take player role (observer only)

```json
{"type": "claim"}
```

Demotes the current player to observer and promotes this client to player. Both clients receive a `role` message.

---

## Running Tests

```bash
# Unit tests (grid logic, match detection, gravity, brick destruction)
make test

# Integration tests (builds headless binary, drives it via TCP + Unix socket)
make test-integration
```

`make test` runs `go test -count=1 ./...` (excludes the `integration` build tag).  
`make test-integration` runs `go test -tags integration -count=1 -v -timeout 90s .`,  
which builds a temporary headless binary and drives it end-to-end.

Integration test coverage includes: TCP connect, Unix connect, second-client observer,
observer claim, invalid move, valid move, cascade detection, fall animation ball data,
Ball Attack game-over, and mode switching.

---

## Files

```
ballfall/
  main.go              entry point: CLI flags, start servers + ebiten/headless loop
  game.go              Game struct, LogicTick(), Update(), state machine, dropBall()
  grid.go              Grid, FindMatches, FindMatchGroups, ApplyGravity,
                       GravityAndFill (Demo), GravityOnly (Attack),
                       Swap, NewGrid, NewAttackGrid, DestroyAdjacentBricks
  draw.go              ebiten Draw(), Layout(), mode-select HUD, game-over overlay,
                       drag visual (ghost ring, destination highlight, lifted ball)
  mouse_gui.go         ebiten mouse handler (drag-to-swap, mode button clicks)
  mouse_nogui.go       stub for -tags nogui builds
  gui.go               runGUI(): ebiten.RunGame + independent 60 Hz LogicTick goroutine
  gui_nogui.go         runGUI() stub for -tags nogui builds
  protocol.go          all shared message types
  hub.go               Hub: client registry, Broadcast, DispatchMove, ClaimPlayer
  server_tcp.go        TCP (:7777) and Unix socket listeners + per-conn read/write
  server_ws.go         WebSocket (/ws) + HTTP static server (/)
  web.go               web assets embedded via go:embed
  grid_test.go         unit tests: match detection, gravity, cascade, bricks
  integration_test.go  integration tests (build tag: integration)
  cmd/connect3/
    main.go            test player: connects via TCP, plays match-creating swaps
  web/
    index.html         browser UI: canvas renderer, WebSocket, drag-to-swap,
                       swap/fall/flash animations, game-over overlay
  docs/
    game-specification.md  (this file)
    tasks.md
    conventions.md
  ballfall.service     systemd unit (headless mode, RuntimeDirectory=ballfall)
  Makefile
  go.mod / go.sum
```
