# Whale

## Running

### Run with Redpanda

```sh
docker compose up
```

| Service | Description |
|---|---|
| **whale** | Built from the local `Dockerfile`, port `8080`, SQLite persisted in a named volume |
| **redpanda** | Single-node dev container, Kafka API exposed on `19092` (external), with a healthcheck |
| **redpanda-console** | Web UI at http://localhost:8088 |

### Build and run in Docker + test

```sh
make docker-test
```
