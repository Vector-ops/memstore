package protocol

const (
	CommandSET  = "SET"
	CommandGET  = "GET"
	CommandKEYS = "KEYS"
)

type Command interface{}

type SetCommand struct {
	Key, Value []byte
}
type GetCommand struct {
	Key []byte
}

type KeysCommand struct{}
