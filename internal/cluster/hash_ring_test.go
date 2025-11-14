package cluster

import (
	"log"
	"testing"

	"github.com/vector-ops/memstore/internal/transport"
)

func TestHashRing(t *testing.T) {
	hashRing := NewHashRing()

	node1 := transport.NewMockTransport()
	node2 := transport.NewMockTransport()
	node3 := transport.NewMockTransport()

	log.Printf("----------Ping Commands-------")
	hashRing.AddNode(node1)
	hashRing.AddNode(node2)
	hashRing.AddNode(node3)

	log.Printf("----------Set Commands-------")
	hashRing.StoreKey("animal", []byte("345"))
	hashRing.StoreKey("zebra", []byte("123"))
	hashRing.StoreKey("tiger", []byte("567"))
	hashRing.StoreKey("xenophone", []byte("890"))
	hashRing.StoreKey("dinosaur", []byte("345"))
	hashRing.StoreKey("help", []byte("123"))
	hashRing.StoreKey("look", []byte("567"))
	hashRing.StoreKey("peek", []byte("890"))
	hashRing.StoreKey("plague", []byte("345"))
	hashRing.StoreKey("zen", []byte("123"))
	hashRing.StoreKey("telephone", []byte("567"))
	hashRing.StoreKey("xmas", []byte("890"))

	log.Printf("----------Get Commands-------")
	hashRing.RetrieveKey("animal")
	hashRing.RetrieveKey("zebra")
	hashRing.RetrieveKey("tiger")
	hashRing.RetrieveKey("xenophone")
	hashRing.RetrieveKey("dinosaur")
	hashRing.RetrieveKey("help")
	hashRing.RetrieveKey("look")
	hashRing.RetrieveKey("peek")
	hashRing.RetrieveKey("plague")
	hashRing.RetrieveKey("zen")
	hashRing.RetrieveKey("telephone")
	hashRing.RetrieveKey("xmas")
}
