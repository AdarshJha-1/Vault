package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AdarshJha-1/Vault/internal/store"
	"github.com/AdarshJha-1/Vault/internal/wal"
)

func handlePing() string {
	return "+Pong\r\n"
}

func handleSet(tokens []string, storage store.Store) string {
	if len(tokens) < 3 {
		return "-Error: SET requires key and value\r\n"
	}
	storage.Set(tokens[1], tokens[2])
	return "+OK\r\n"
}

func handleGet(tokens []string, storage store.Store) string {
	if len(tokens) < 2 {
		return "-Error: GET requires a key\r\n"
	}
	value, ok := storage.Get(tokens[1])
	if !ok {
		return "$-1\r\n"
	}
	return fmt.Sprintf("$ %s \r\n %s \r\n", strconv.Itoa(len(value)), value)
}

func handleDelete(tokens []string, storage store.Store) string {
	if len(tokens) < 2 {
		return "-Error: DEL requires a key\r\n"
	}
	ok := storage.Delete(tokens[1])
	if !ok {
		return "$-1\r\n"
	}
	return "+OK\r\n"
}

func ProcessCommand(commandArg string, storage store.Store, wal wal.WAL) string {
	tokens := strings.Fields(commandArg)
	if len(tokens) == 0 {
		return "-Error: Empty command\r\n"
	}

	cmd := strings.ToUpper(tokens[0])

	switch cmd {
	case "PING":
		return handlePing()
	case "SET":
		if err := wal.WriteEntry(commandArg); err != nil {
			return "-ERR WAL write failed\r\n"
		}
		return handleSet(tokens, storage)
	case "GET":
		return handleGet(tokens, storage)
	case "DEL":
		if err := wal.WriteEntry(commandArg); err != nil {
			return "-ERR WAL write failed\r\n"
		}
		return handleDelete(tokens, storage)
	default:
		return "-Error: Unknown command\r\n"
	}
}
