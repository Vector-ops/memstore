package main

import (
	"flag"
	"log"
)

const DefaultListenAddr = ":3000"

func main() {
	var listenAddr string
	flag.StringVar(&listenAddr, "listenAddr", DefaultListenAddr, "Specify the listen address for the memstore server or use default (3000)")
	flag.Parse()
	server := NewServer(Config{
		ListenAddr: listenAddr,
	})

	log.Fatal(server.Start())
}
