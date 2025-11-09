package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/vector-ops/memstore/internal/config"
	"github.com/vector-ops/memstore/internal/server"
)

func main() {
	var nodeId string
	flag.StringVar(&nodeId, "node-id", "", "Specify the node id to start")

	var configPath string
	flag.StringVar(&configPath, "config", "config.json", "Path to config file")
	flag.Parse()

	if nodeId == "" {
		log.Fatalf("-node-id not provided")
	}

	f, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}

	var nodes []config.NodeConfig
	if err = json.NewDecoder(f).Decode(&nodes); err != nil {
		log.Fatal(err)
	}

	var serverConfig config.NodeConfig
	replicas := make([]config.NodeConfig, 0)
	for _, n := range nodes {
		if n.Id == nodeId {
			serverConfig = n
		}
	}

	if serverConfig.Id == "" {
		log.Fatalf("node config not found")
	}

	for _, n := range nodes {
		if n.Id != nodeId {
			if serverConfig.Role == "master" {
				replicas = append(replicas, n)
			}
		}
	}

	server := server.NewServer(server.Config{
		NodeConfig: serverConfig,
		Replicas:   replicas,
	})

	log.Fatal(server.Start())
}
