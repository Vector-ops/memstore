package client

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client, err := New("localhost:5000")
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second)
	for i := 0; i < 10; i++ {
		if err := client.Set(context.Background(), fmt.Sprintf("sh%dt", i), fmt.Sprintf("T%dn", i)); err != nil {
			log.Fatal(err)
		}
		val, err := client.Get(context.Background(), fmt.Sprintf("sh%dt", i))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Heres it: ", val)
	}

}
