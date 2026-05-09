# Whale

## Running

### Run with Redpanda

```sh
docker compose up --build
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

## Kafka interaction

Whale listens on the `people` Kafka topic (configurable via `APP_KAFKA_TOPIC`). Messages are JSON in the same format as the HTTP `/save` endpoint.

### 1. Start the stack

```sh
docker compose up --build
```

Wait until you see in the logs:
```
whale-1  | kafka consumer started brokers=[redpanda:9092] topic=people group=whale
```

### 2. Post a message via the test script

```sh
./test_kafka.sh
```

The script:
- Produces a JSON person record to the `people` topic via Redpanda's HTTP Proxy (Pandaproxy) on `localhost:18082`
- Waits 2 seconds for whale to consume it
- Fetches the stored person back from the whale HTTP API on `localhost:8080`

Expected output:
```
Posting person to Kafka topic 'people' via Pandaproxy...
  external_id: 3f2b1c4a-...
{"offsets":[{"partition":0,"offset":0}]}

Waiting for whale to consume and store the record...
Fetching person from whale API...
{
  "external_id": "3f2b1c4a-...",
  "name": "Test Person",
  "email": "test@example.com",
  "date_of_birth": "1990-01-01T00:00:00Z"
}
```

Whale logs will show:
```
whale-1  | kafka: received message topic=people partition=0 offset=0
whale-1  | kafka: saved person external_id=3f2b1c4a-...
```

### 3. Inspect in Redpanda Console

Open http://localhost:8088/overview in a browser.

- **Topics → people** — browse produced messages and their offsets
- **Consumer Groups → whale** — see the committed offset and lag for the whale consumer
- **Brokers** — cluster health and partition assignments

### Configuration

| Environment variable | Default | Description |
|---|---|---|
| `APP_KAFKA_BROKERS` | `localhost:9092` | Comma-separated broker list |
| `APP_KAFKA_TOPIC` | `people` | Topic to consume from |
| `APP_KAFKA_GROUP` | `whale` | Consumer group ID |
