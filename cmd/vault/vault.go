package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/AdarshJha-1/Vault/internal/server"
	"github.com/AdarshJha-1/Vault/internal/store"
	"github.com/AdarshJha-1/Vault/internal/wal"
)

const (
	walLogDir = "data/wal"
)

func graceFullShutdown(done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl-C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done <- true

}

func main() {
	port := 5555
	args := os.Args
	if len(args) >= 2 {
		var err error
		port, err = strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Conversion error:", err)
			os.Exit(1)
		}
	}

	storage := store.GetStore()

	newWal, err := wal.OpenWAL(walLogDir, 16*1024*1024, 10)
	if err != nil {
		log.Fatal("Error::", err)
	}

	err = newWal.LoadToVault(storage)
	if err != nil {
		log.Fatal("Error::", err)
	}

	srv := server.GetVaultServer(port, storage, newWal)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	go srv.Run()

	<-done
	fmt.Printf("Vault server shutting down\n")
	srv.ShutDown()
}
