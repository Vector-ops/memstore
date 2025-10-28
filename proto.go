package main

const (
	CommandSET = "SET"
	CommandGET = "GET"
)

type Command interface{}

type SetCommand struct {
	key, value []byte
}
type GetCommand struct {
	key []byte
}
