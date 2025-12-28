package main

import (
	"flag"
	"log"
	"time"

	"github.com/vector-ops/memstore/internal/config"
	"github.com/vector-ops/memstore/internal/server"
)

func main() {
	var nodeId string
	flag.StringVar(&nodeId, "node-id", "", "Specify the node id to start")

	// var configPath string
	// flag.StringVar(&configPath, "config", "", "Path to config file")
	// flag.Parse()

	var tcpAddr string
	flag.StringVar(&tcpAddr, "tcp", "", "Specify the tcp address for the server")

	var leaderAddr string
	flag.StringVar(&leaderAddr, "leader", "", "Specify the leader address, leave blank if leader")

	var replicas int
	flag.IntVar(&replicas, "replicas", 0, "Specify the no. of replicas")

	var replicaTimeout int
	flag.IntVar(&replicaTimeout, "replica-timeout", 0, "Specify the replica connect timeout in milliseconds")

	var replicationMaxRetries int
	flag.IntVar(&replicationMaxRetries, "max-retries", 0, "Specify the max replication retries")

	var pollInterval int
	flag.IntVar(&pollInterval, "poll", 0, "Specify the leader poll interval in milliseconds")

	cfg := config.ServerConfig{
		Id:                    nodeId,
		TCPAddr:               tcpAddr,
		LeaderAddr:            leaderAddr,
		Replicas:              replicas,
		ReplicationTimeout:    time.Duration(replicaTimeout) * time.Millisecond,
		ReplicationMaxRetries: replicationMaxRetries,
		PollInterval:          time.Duration(pollInterval) * time.Millisecond,
	}

	// if configPath != "" {
	// 	cfg = parseConfig(configPath)
	// } else {

	// }

	server := server.NewServer(cfg)

	log.Fatal(server.Start())
}

// func parseConfig(p string) config.ServerConfig {

// 	var cfg config.ServerConfig

// 	f, err := os.Open(p)
// 	if err != nil {
// 		panic(err)
// 	}

// 	b, err := io.ReadAll(f)
// 	if err != nil {
// 		panic(err)
// 	}

// 	_ = b

// 	return cfg

// }
