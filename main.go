package main

import (
	"log"
	"net"

	"github.com/ldryt/mcpulse/config"
	"github.com/ldryt/mcpulse/gateway"
)

func main() {
	log.Println("Loading configuration...")
	err := config.Load()
	if err != nil {
		log.Fatalln("Error loading configuration:", err)
	}

	gw := gateway.New(config.Get())

	address := config.Get().ListenAddress
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		log.Fatalf("Error creating listener on %v: %v", address, err)
	}
	defer listener.Close()

	log.Printf("Listening on %v", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection on %v: %v", listener.Addr(), err)
			continue
		}

		go gw.HandleConnection(conn)
	}
}
