package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/AdarshJha-1/Vault/internal/server"
	"github.com/AdarshJha-1/Vault/internal/store"
)

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
	srv := server.GetVaultServer(port, storage)
	srv.Run()
	defer srv.ShutDown()
}
