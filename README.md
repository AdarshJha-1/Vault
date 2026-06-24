# Vault

Vault is an in-memory key-value store written in Go with persistence through a Write-Ahead Log (WAL).

The project is inspired by systems such as Redis and focuses on understanding how storage engines, persistence, crash recovery, and concurrent data structures work internally.

## Features

### Key-Value Operations

Supported commands:

```text
PING
SET <key> <value>
GET <key>
DEL <key>
```

Example:

```text
SET name adarsh
GET name
DEL name
PING
```

---

### Concurrent In-Memory Store

- Thread-safe key-value storage
- Uses `sync.RWMutex`
- Multiple readers can access data concurrently
- Writers acquire exclusive access

---

### TCP Server

- Custom TCP server implementation
- Handles multiple client connections concurrently using goroutines
- Text-based command protocol

Example:

```bash
nc localhost 5555
```

---

### Write-Ahead Logging (WAL)

Every command is persisted before execution.

WAL entry format:

```text
+------------+------------------+
| Length (4) | Protobuf Payload |
+------------+------------------+
```

Each entry contains:

```protobuf
message WALEntry {
    uint64 lsn = 1;
    string command = 2;
    uint32 crc = 3;
}
```

Fields:

| Field   | Purpose              |
| ------- | -------------------- |
| LSN     | Log Sequence Number  |
| Command | Original command     |
| CRC     | Corruption detection |

---

### Crash Recovery

On startup Vault:

1. Opens WAL directory
2. Loads all WAL segments
3. Validates CRC of every entry
4. Replays valid commands
5. Rebuilds in-memory state
6. Restores latest LSN

This allows data to survive process crashes and restarts.

---

### WAL Segmentation

Logs are split into segment files.

Example:

```text
data/
└── wal/
    ├── segment-0
    ├── segment-1
    ├── segment-2
```

Benefits:

- Prevents single huge log file
- Faster startup recovery
- Easier log management

---

### WAL Rotation

When a segment exceeds the configured size:

```text
Current Segment Full
        ↓
Create New Segment
        ↓
Continue Writing
```

Old segments are removed when the configured maximum segment count is reached.

---

### Corruption Detection

Each WAL entry stores a CRC32 checksum.

During recovery:

```text
Read Entry
    ↓
Verify CRC
    ↓
Valid ? Replay : Stop
```

This prevents replaying corrupted or partially written entries.

---

### Recovery from Partial Writes

If Vault crashes while writing:

```text
[length]
[partial protobuf data]
```

Recovery detects:

```go
io.EOF
io.ErrUnexpectedEOF
```

and safely stops reading the damaged portion of the log.

---

## Project Structure

```text
Vault
├── cmd/
│   └── vault/
│       └── vault.go
│
├── internal/
│   ├── handler/
│   ├── server/
│   ├── store/
│   └── wal/
│
├── proto/
│   └── wal/
│
└── data/
    └── wal/
```

---

## Build

```bash
go build -o bin/vault ./cmd/vault
```

---

## Run

```bash
./bin/vault
```

or

```bash
go run ./cmd/vault
```

Custom port:

```bash
go run ./cmd/vault 8080
```

---

## Current Status

Implemented:

- [x] TCP Server
- [x] Concurrent Key-Value Store
- [x] WAL Persistence
- [x] Protobuf WAL Entries
- [x] CRC Validation
- [x] WAL Replay
- [x] Crash Recovery
- [x] Segment Rotation
- [x] Multi-Client Support

Planned:

- [ ] LRU Cache Eviction
- [ ] Snapshotting
- [ ] WAL Compaction
- [ ] TTL Expiration
- [ ] Benchmarking
- [ ] Unit Tests
- [ ] RESP Protocol Support
