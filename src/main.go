package main

import (
	"flag"
	"log"
	"net"
)

var GlobalConfig Config

func main() {
	configPathPTR := flag.String("config", "./mcpulse.yml", "a path to the configuration file")
	flag.Parse()
	configPath := *configPathPTR

	log.Println("Loading configuration...")
	GlobalConfig, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalln("An error occurred while loading configuration:", err)
	}

	go listenTCP4("SLP", GlobalConfig.SLP.ListenAddress)
}

func listenTCP4(name string, address string) {
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		log.Fatalf("An error occurred while creating listener '%v' (%v): %v\n", name, address, err)
	}
	defer listener.Close()

	log.Printf("Server '%v' listening on %v\n", name, listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("An error occurred while accepting connection:", err)
			continue
		}

		go HandleClient(conn)
	}
}
