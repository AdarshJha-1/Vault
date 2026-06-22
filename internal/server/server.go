package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/AdarshJha-1/Vault/internal/handler"
	"github.com/AdarshJha-1/Vault/internal/store"
	"github.com/AdarshJha-1/Vault/internal/wal"
)

type Server interface {
	Run()
	ShutDown()
}

type vaultServer struct {
	port     int
	listener net.Listener
	running  bool
	storage  store.Store
	wal      wal.WAL
}

func GetVaultServer(port int, storage store.Store, wal wal.WAL) Server {
	return &vaultServer{
		port:    port,
		running: true,
		storage: storage,
		wal:     wal,
	}
}

func (vs *vaultServer) Run() {

	var err error
	vs.listener, err = net.Listen("tcp", ":"+strconv.Itoa(vs.port))
	if err != nil {
		log.Fatal("Error listening:", err)
	}
	defer vs.listener.Close()

	fmt.Printf("Vault server running on port: %d\n", vs.port)

	for vs.running {
		conn, err := vs.listener.Accept()
		if err != nil {
			log.Println("Error accepting conn:", err)
			continue
		}
		go handleConnection(conn, vs.storage, vs.wal)
	}

	fmt.Println("Vault server is shutting down")
}

func (vs *vaultServer) ShutDown() {
	fmt.Printf("Vault server shutting down\n")
}

func handleConnection(conn net.Conn, storage store.Store, wal wal.WAL) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		commandArg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Client disconnected\n")
			} else {
				log.Printf("Read error: %v", err)
			}
			return
		}
		res := handler.ProcessCommand(commandArg, storage, wal)
		if _, err := conn.Write([]byte(res)); err != nil {
			log.Printf("Server write error: %v", err)
		}
	}
}
