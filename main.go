package main

import (
	"log"
	"net"

	"github.com/ldryt/mcpulse/config"
	"github.com/ldryt/mcpulse/pulser"
	"github.com/ldryt/mcpulse/slp"
)

func main() {
	log.Println("Loading configuration...")
	err := config.Load()
	if err != nil {
		log.Fatalln("Error loading configuration:", err)
	}

	go listenTCP4("ServerListPing", config.Get().SLP.ListenAddress, slp.HandleConnection)
	go listenTCP4("Pulser", config.Get().Pulser.ListenAddress, pulser.HandleConnection)

	select {}
}

func listenTCP4(name string, address string, handler func(net.Conn)) {
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		log.Fatalf("Error creating %v listener on %v: %v", name, address, err)
	}
	defer listener.Close()

	log.Printf("%v listening on %v", name, listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection on %v: %v", listener.Addr(), err)
			continue
		}

		go handler(conn)
	}
}
