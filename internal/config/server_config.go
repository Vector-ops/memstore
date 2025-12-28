package config

import "time"

type ServerConfig struct {
	Id          string
	ServerAddr  string
	ClusterAddr string
	LeaderAddr  string

	// leader config
	HeartbeatInterval     time.Duration
	ReplicationTimeout    time.Duration
	ReplicationMaxRetries int
	ReplicaCount          int

	// follower config
	HeartbeatTimeout time.Duration
}
