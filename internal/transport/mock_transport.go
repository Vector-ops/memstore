package transport

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/tidwall/resp"
	"github.com/vector-ops/memstore/internal/protocol"
)

type MockTransport struct {
	id          string
	localStore  map[string][]byte
	prevCommand protocol.Command
	prevKey     string
}

func NewMockTransport() Transport {
	return &MockTransport{
		id:         strconv.Itoa(rand.Int()),
		localStore: make(map[string][]byte),
	}
}

func (t *MockTransport) Ping() (string, error) {
	return t.id, nil
}

func (t *MockTransport) Send(msg []byte) (int, error) {
	buf := bytes.NewReader(msg)
	rd := resp.NewReader(buf)

	v, _, err := rd.ReadValue()
	if err != nil {
		return 0, err
	}

	if v.Type() == resp.Array {
		for _, c := range v.Array() {
			switch strings.ToUpper(c.String()) {
			case protocol.CommandGET:
				if len(v.Array()) != 2 {
					return 0, fmt.Errorf("invalid number of variables for GET command")
				}
				cmd := protocol.GetCommand{
					Key: v.Array()[1].Bytes(),
				}
				t.prevCommand = protocol.CommandGET
				t.prevKey = string(cmd.Key)
				log.Printf("node %s, recieved send request for command Get: %s", t.id, string(cmd.Key))
			case protocol.CommandSET:
				if len(v.Array()) != 3 {
					return 0, fmt.Errorf("invalid number of variables for SET command")
				}
				cmd := protocol.SetCommand{
					Key:   v.Array()[1].Bytes(),
					Value: v.Array()[2].Bytes(),
				}
				t.localStore[string(cmd.Key)] = cmd.Value
				t.prevCommand = protocol.CommandSET
				t.prevKey = string(cmd.Key)
				log.Printf("node %s, recieved send request for command Set: %s:%s", t.id, string(cmd.Key), string(cmd.Value))
			default:
			}
		}
	}
	return 0, nil
}

func (t *MockTransport) Read() ([]byte, error) {
	if t.prevCommand == protocol.CommandGET {
		return t.localStore[t.prevKey], nil
	}
	return nil, nil
}

func (t *MockTransport) ReadLoop() error {
	return nil
}

func (t *MockTransport) GetRemoteAddress() string {
	return "mock-address"
}

func (t *MockTransport) GetLocalAddress() string {
	return "mock-address"
}
