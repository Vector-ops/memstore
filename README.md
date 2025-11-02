# Memstore: Distributed key value store

## Goals
- [ ] Add partitioning logic
- [ ] Add replication protocol (gossip, consensus)
- [ ] Add http gateway


## Plan
- [x] Static config for starting servers
- [ ] Add retry logic for nodes present in the static config

## Master selection
- Need consensus like raft or paxos to avoid disagreement among replicas
