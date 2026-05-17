# Development Conventions

## Directory layout for artifacts

| Path | Purpose |
|------|---------|
| `tmp/` | **Scratch area for temporary development/debug artifacts**: test binaries, one-off scripts, intermediate outputs. Gitignored. Created by `make linux-amd64` etc. |
| `/run/ballfall/` | **Runtime socket** — created by systemd `RuntimeDirectory=ballfall`, cleaned up on service stop. Contains `ballfall.sock`. |
| `/opt/ballfall/` | **Production install** — binary lives here, managed by `make install` / `make deploy`. |

## Socket file placement

Unix sockets go in `/run/<service>/` — the conventional location for sockets and PID files on modern Linux (tmpfs, auto-cleaned on reboot). The systemd service declares `RuntimeDirectory=ballfall`, which creates `/run/ballfall/` with correct permissions before `ExecStart`.

For **local development** (no root, `/run/ballfall/` unavailable), override the default:

```bash
go run . --unix /tmp/ballfall.sock    # or any writable path
```

The binary calls `os.MkdirAll` on the socket's parent directory as a best-effort fallback; `StartUnix` logs and continues if it fails, leaving TCP (`:7777`) and HTTP/WebSocket (`:8080`) still functional.

## Makefile targets

```
make build        # local macOS binary (./ballfall)
make linux-amd64  # cross-compile → tmp/ballfall-linux-amd64
make deploy       # linux-amd64 + scp + systemctl restart on ora-m
make install      # first-time: deploy + install systemd unit + enable
make status       # ssh ora-m, systemctl status ballfall
make logs         # ssh ora-m, journalctl -u ballfall -f
make clean        # rm -rf tmp/ ballfall
```

## tmp/ usage

`tmp/` is a scratch pad — write test scripts, debug binaries, and intermediate files here without polluting the repo. Examples:

```bash
# build a local test binary
go build -o tmp/ballfall-local .

# quick manual socket test
echo '{"type":"move","row":0,"col":0,"dir":"right"}' | nc -w1 localhost 7777
```

Nothing in `tmp/` is committed; the directory is created on demand by the Makefile.
