# Vault

Vault is an in-memory key-value store written in Go.

I built this project to understand how systems like Redis work internally instead of treating them as a black box. The focus was learning storage engines, persistence, crash recovery, caching, concurrency, and networking by implementing them from scratch.

---

## Features

### Supported Commands

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

## In-Memory Store

Vault stores data completely in memory.

Current implementation:

- Thread-safe using `sync.RWMutex`
- O(1) GET, SET and DELETE
- Hash map + doubly linked list
- LRU (Least Recently Used) eviction
- Configurable maximum capacity

The LRU cache keeps recently accessed keys at the front of the list. Once the configured capacity is reached, the least recently used key is automatically evicted.

---

## TCP Server

The server is built using Go's `net` package.

Features:

- Multiple client support
- One goroutine per connection
- Simple text-based command protocol
- Graceful shutdown

Example:

```bash
nc localhost 5555
```

---

## Write-Ahead Log (WAL)

Every write operation is written to disk before modifying memory.

Each WAL entry contains:

```protobuf
message WALEntry {
    uint64 lsn = 1;
    string command = 2;
    uint32 crc = 3;
}
```

Fields:

| Field   | Description                             |
| ------- | --------------------------------------- |
| LSN     | Log Sequence Number                     |
| Command | Original command                        |
| CRC     | CRC32 checksum for corruption detection |

Protocol Buffers are used to serialize WAL entries.

---

## WAL Segments

Instead of storing every log in one file, Vault splits logs into segments.

Example:

```text
data/
└── wal/
    ├── segment-0
    ├── segment-1
    ├── segment-2
```

Benefits:

- Prevents one huge log file
- Faster startup
- Easier log management
- Automatic log rotation

---

## Crash Recovery

On startup Vault:

1. Opens every WAL segment
2. Reads entries in order
3. Verifies CRC
4. Stops if corruption is detected
5. Replays valid commands
6. Restores the latest LSN

This allows data to survive crashes and process restarts.

---

## Corruption Detection

Every WAL entry stores a CRC32 checksum.

During recovery:

```text
Read Entry
     │
     ▼
Verify CRC
     │
 ┌───┴────┐
 │        │
Valid   Invalid
 │        │
 ▼        ▼
Replay   Stop Recovery
```

If the server crashes while writing an entry, recovery safely stops when it encounters:

- `io.EOF`
- `io.ErrUnexpectedEOF`

This prevents replaying partially written data.

---

## Benchmarks

Current benchmarks for the in-memory store:

```
BenchmarkSetKV      ~482 ns/op
BenchmarkGetKV      ~163 ns/op
BenchmarkDeleteKV   ~234 ns/op
```

---

## Project Structure

```text
Vault
├── cmd/
│   └── vault/
│       └── main.go
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
├── data/
│   └── wal/
│
├── Makefile
├── README.md
├── go.mod
└── go.sum
```

---

## Build

```bash
make build
```

or

```bash
go build -o bin/vault ./cmd/vault
```

---

## Run

```bash
make run
```

or

```bash
./bin/vault
```

Custom port:

```bash
go run ./cmd/vault 8080
```

---

## Current Status

Implemented:

- [x] Concurrent In-Memory Store
- [x] LRU Cache
- [x] TCP Server
- [x] Multi-client Support
- [x] Write-Ahead Log (WAL)
- [x] WAL Replay
- [x] Crash Recovery
- [x] WAL Segmentation
- [x] WAL Rotation
- [x] CRC32 Validation
- [x] Graceful Shutdown
- [x] Benchmarks

Planned:

- [ ] Snapshotting
- [ ] WAL Compaction
- [ ] TTL Expiration
- [ ] RESP Protocol
- [ ] Unit Tests

---

## Why I Built This

The goal of Vault wasn't to build another Redis clone. I wanted to understand how a storage engine actually works by implementing the core pieces myself.

Through this project I learned about:

- TCP servers in Go
- Concurrent programming
- LRU cache design
- Write-Ahead Logging
- Crash recovery
- Data serialization with Protocol Buffers
- CRC based corruption detection
- Log segmentation and rotation
- Basic storage engine design

The project is still growing as I continue learning more about database internals and distributed systems.
