# Ball Fall — Game Specification

## Overview

A match-3 gravity game. A rectangular grid is filled with colored balls. A player swaps adjacent balls; when 3 or more of the same color align (or form a 2×2 block) they disappear, remaining balls fall due to gravity, and new random balls fill the top. Cascade chains score bonus points.

## Grid

| Parameter | Value |
|-----------|-------|
| Width     | 8 columns |
| Height    | 10 rows |
| Cell size | 60 px (display) |
| Colors    | Red (1), Green (2), Blue (3), Yellow (4) |
| Empty     | 0 |

Row 0 is the top. New balls appear at row 0 and fall downward.

## Match Rules

Matches are detected after every swap (and after every cascade):

1. **Horizontal run**: 3 or more consecutive same-color balls in a row.
2. **Vertical run**: 3 or more consecutive same-color balls in a column.
3. **2×2 block**: any 2×2 area where all four cells share the same color.

All matching cells across all rules are collected into one set and cleared simultaneously.

## Game Flow

```
stateWaiting
  ↓ receive move
stateProcessing (swap + immediate match check)
  → no match found → swap reverted → stateWaiting (board broadcast)
  → matches found  → stateClearing
stateClearing (30-frame flash animation)
  ↓ timer expires → remove matched cells → apply gravity → fill empty cells
stateChecking (cascade check)
  → new matches found → stateClearing
  → no matches       → stateWaiting (board broadcast)
```

## Player Moves

A move specifies a cell by `(row, col)` and a direction (`left`, `right`, `up`, `down`). The game swaps the selected cell with its neighbor in that direction. If the resulting board has no matches the swap is immediately reversed (invalid move); the board is broadcast with an `"invalid"` event flag.

Only one player can be active at a time. The first client to connect in `player` role becomes the active player. If that client disconnects, the next client to claim the `player` role is promoted.

## Scoring

| Event | Points |
|-------|--------|
| Each ball cleared | 1 |
| Cascade multiplier | ×cascade_depth |

## Clients and Protocols

The game binary runs as a server. Multiple clients can connect simultaneously.

### Client Roles

| Role | Can send moves | Receives board |
|------|---------------|----------------|
| `player` | yes (first connected) | yes |
| `observer` | no | yes |

### Transports

| Transport | Address | Notes |
|-----------|---------|-------|
| TCP socket | `:7777` | newline-delimited JSON |
| Unix socket | `/tmp/ballfall.sock` | same protocol |
| WebSocket | `ws://localhost:8080/ws` | same JSON protocol |
| HTTP | `http://localhost:8080/` | serves browser UI |

### Message Protocol (JSON, newline-delimited)

**Server → Client: board state** (broadcast whenever the board settles or a significant animation begins)

```json
{
  "type":    "state",
  "board":   [[1,2,3,4,2,1,3,4], ...],
  "width":   8,
  "height":  10,
  "score":   150,
  "status":  "waiting",
  "matches": [[3,4],[3,5],[3,6]],
  "cascade": 0
}
```

`status` values: `"waiting"` | `"clearing"` | `"invalid"`

**Server → Client: welcome** (on connect)

```json
{"type": "welcome", "role": "player", "width": 8, "height": 10}
```

**Client → Server: move** (player only)

```json
{"type": "move", "row": 3, "col": 4, "dir": "right"}
```

`dir` values: `"left"` | `"right"` | `"up"` | `"down"`

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
  main.go          entry point: start all servers + ebiten
  game.go          Game struct, Update(), Draw()
  grid.go          Grid, FindMatches, ApplyGravity, FillEmpty
  protocol.go      shared message types
  hub.go           Hub: client registry, broadcast, move dispatch
  server_tcp.go    TCP and Unix socket listeners
  server_ws.go     WebSocket + HTTP server
  cmd/player/
    main.go        test player: connects via TCP, plays randomly
  web/
    index.html     browser UI (vanilla JS + WebSocket)
  docs/
    game-specification.md  (this file)
    tasks.md
  go.mod
  go.sum
```
