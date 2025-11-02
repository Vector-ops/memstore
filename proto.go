package main

const (
	CommandSET  = "SET"
	CommandGET  = "GET"
	CommandKEYS = "KEYS"
)

type Command interface{}

type SetCommand struct {
	key, value []byte
}
type GetCommand struct {
	key []byte
}

type KeysCommand struct{}
