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

	listener, err := net.Listen("tcp4", GlobalConfig.ListenAddress)
	if err != nil {
		log.Fatalln("An error occurred while creating listener:", err)
	}
	defer listener.Close()

	log.Println("Server listening on", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("An error occurred while accepting connection:", err)
			continue
		}

		go HandleClient(conn)
	}
}
