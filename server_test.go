package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/vector-ops/goredis/client"
)

func TestServerWithClients(t *testing.T) {
	server := NewServer(Config{})
	go func() {
		log.Fatal(server.Start())
	}()
	time.Sleep(time.Second)
	nclients := 10
	wg := sync.WaitGroup{}
	wg.Add(nclients)
	for i := 0; i < nclients; i++ {
		go func(it int) {
			client, err := client.New("localhost:3000")
			if err != nil {
				log.Fatal(err)
			}
			defer client.Close()

			key := fmt.Sprintf("foo_%d", it)
			value := fmt.Sprintf("bar_%d", it)

			if err := client.Set(context.Background(), key, value); err != nil {
				log.Fatal(err)
			}
			val, err := client.Get(context.Background(), key)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("key foo_%d => %s\n", it, val)
			wg.Done()
		}(i)
	}

	wg.Wait()

	time.Sleep(time.Second)
	if len(server.peers) != 0 {
		t.Fatalf("expected 0 peers but got %d", len(server.peers))
	}
}
