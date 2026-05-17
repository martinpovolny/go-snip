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

Endless match-3. The grid starts fully populated. No game-over condition. Score accumulates indefinitely.

### Ball Attack

A survival mode. Balls fall from the top at a fixed rate; the player must clear them by swapping to form matches. The game ends when a column is too full to accept the next incoming ball.

**Starting state:** the bottom 2 rows (rows 8–9) are pre-filled with random balls; rows 0–7 are empty.

**Drop mechanic (every 1 second while the game is running):**

1. A ball of random color is chosen.
2. A column is chosen at random.
3. The ball falls from row 0 to the first occupied row below minus one (i.e. it lands on top of the current stack in that column).
4. **Slide rule:** if the landing cell is occupied (the column has no empty space to receive the ball — the stack reaches row 0), the game is over (see Game Over below). Otherwise if the ball lands directly on top of an isolated ball (no horizontal neighbor on either side at the landing row), it slides one step randomly left or right; if the slide destination is also full it slides the other way; if both are full the ball stays put and may trigger a match normally.
5. After the ball is placed, the match-3 clearing logic runs exactly as in Demo (find groups → clear → gravity → cascade). Clears do not pause the drop timer.

**Game over:** triggered when the next drop would land at row 0 or above (the chosen column is completely full). A game-over state is broadcast; the board freezes and no further drops or player moves are accepted until the player re-selects a mode.

**Scoring:** same rules as Demo (1 point per cleared ball, cascade multiplier). Score resets on each new Ball Attack session.

## Match Rules

Matches are detected after every swap (and after every cascade):

1. **Horizontal run**: 3 or more consecutive same-color balls in a row.
2. **Vertical run**: 3 or more consecutive same-color balls in a column.
3. **2×2 block**: any 2×2 area where all four cells share the same color.

All matching cells across all rules are collected into one set and cleared simultaneously.

## Game Flow

### Demo

```
stateWaiting
  ↓ receive move
stateSwapping (swap + immediate match check)
  → no match found → swap reverted → stateWaiting (board broadcast)
  → matches found  → stateClearing
stateClearing (30-frame flash animation)
  ↓ timer expires → remove matched cells → apply gravity → fill empty cells
stateFalling (gravity animation)
  → new matches found → stateClearing (cascade)
  → no matches       → stateWaiting (board broadcast)
```

### Ball Attack

Same swap/clear/fall states as Demo, plus:

```
(every 1 s, regardless of swap state)
  drop ticker fires
  → chosen column full at row 0 → stateGameOver (broadcast game_over)
  → ball placed (with optional slide) → match check → clear/fall/cascade as normal
```

## Player Moves

A move specifies a cell by `(row, col)` and a direction (`left`, `right`, `up`, `down`). The game swaps the selected cell with its neighbor in that direction. If the resulting board has no matches the swap is immediately reversed (invalid move); the board is broadcast with an `"invalid"` event flag.

Moves are ignored in `stateGameOver`.

Only one player can be active at a time. The first client to connect in `player` role becomes the active player. If that client disconnects, the next client to claim the `player` role is promoted.

## Scoring

| Event | Points |
|-------|--------|
| Each ball cleared | 1 |
| Cascade multiplier | ×cascade_depth |

Score resets when a new Ball Attack session starts. Demo score accumulates across mode switches back to Demo.

## Mode Selection

### Web UI

A pair of buttons (or a `<select>`) in the header bar next to the score. The active mode is highlighted. Clicking the inactive mode sends a `set_mode` message and resets the game server-side.

### GUI (ebiten)

Two small toggle buttons rendered in the top margin of the window, always visible. The active mode is highlighted. Clicking the inactive mode button resets and switches. In addition, when Ball Attack reaches game-over, a centered overlay is drawn showing the final score and a prompt; clicking anywhere on the overlay (or pressing Space/Enter) re-selects the same mode and starts a new session.

## Clients and Protocols

The game binary runs as a server. Multiple clients can connect simultaneously.

### Client Roles

| Role | Can send moves | Receives board |
|------|---------------|----------------|
| `player` | yes | yes |
| `observer` | no | yes |

### Transports

| Transport | Default address | Notes |
|-----------|----------------|-------|
| TCP socket | `:7777` | newline-delimited JSON |
| Unix socket | `./ballfall.sock` | same protocol |
| WebSocket | `ws://localhost:8080/ws` | same JSON protocol |
| HTTP | `http://localhost:8080/` | serves browser UI |

### Message Protocol (JSON, newline-delimited)

**Server → Client: board state**

```json
{
  "type":    "state",
  "board":   [[1,2,3,4,2,1,3,4], ...],
  "width":   8,
  "height":  10,
  "score":   150,
  "status":  "waiting",
  "matches": [[3,4],[3,5],[3,6]],
  "cascade": 0,
  "mode":    "demo"
}
```

`status` values: `"waiting"` | `"clearing"` | `"invalid"` | `"gameover"`

`mode` values: `"demo"` | `"attack"`

**Server → Client: welcome** (on connect)

```json
{"type": "welcome", "role": "player", "width": 8, "height": 10, "mode": "demo"}
```

**Server → Client: game over** (Ball Attack only)

```json
{"type": "game_over", "score": 342, "mode": "attack"}
```

**Client → Server: move** (player only, ignored during game-over)

```json
{"type": "move", "row": 3, "col": 4, "dir": "right"}
```

**Client → Server: set mode** (player only)

```json
{"type": "set_mode", "mode": "attack"}
```

Resets the game and starts the chosen mode. Broadcasts a fresh state to all clients.

## Architecture

```
┌─────────────────────────────────────────────────┐
│                  game binary                    │
│                                                 │
│  ┌──────────┐   ┌──────┐   ┌────────────────┐  │
│  │  ebiten  │   │ Grid │   │      Hub       │  │
│  │ Display  │←──│ Game │──→│ (broadcast to  │  │
│  │ (local)  │   │ Logic│   │  all clients)  │  │
│  └──────────┘   └──────┘   └───────┬────────┘  │
│                                    │            │
│              ┌─────────────────────┤            │
│              ↓                     ↓            │
│     ┌──────────────┐   ┌─────────────────────┐ │
│     │  TCP/Unix    │   │  WebSocket + HTTP   │ │
│     │   server     │   │      server         │ │
│     └──────┬───────┘   └──────────┬──────────┘ │
└────────────┼──────────────────────┼────────────┘
             ↓                      ↓
      player program         browser / web UI
      (cmd/player)           (web/index.html)
```

## Files

```
ballfall/
  main.go          entry point: start all servers + ebiten/headless loop
  game.go          Game struct, Update(), state machine
  grid.go          Grid, FindMatches, ApplyGravity, FillEmpty, Swap, NewGrid
  draw.go          ebiten Draw(), Layout(), mode-select HUD, game-over overlay
  mouse_gui.go     ebiten mouse handler (drag-to-swap, mode button clicks)
  mouse_nogui.go   stub for -tags nogui builds
  gui.go           runGUI() — ebiten.RunGame wrapper
  gui_nogui.go     runGUI() stub for -tags nogui builds
  protocol.go      shared message types
  hub.go           Hub: client registry, broadcast, move dispatch
  server_tcp.go    TCP and Unix socket listeners
  server_ws.go     WebSocket + HTTP server
  web.go           web assets embedded via go:embed
  cmd/connect3/
    main.go        test player: connects via TCP, plays match-creating swaps
  web/
    index.html     browser UI (vanilla JS + WebSocket)
  docs/
    game-specification.md  (this file)
    tasks.md
    conventions.md
  ballfall.service systemd unit
  Makefile
  go.mod / go.sum
```
