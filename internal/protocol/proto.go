package protocol

const (
	CommandSET  = "SET"
	CommandGET  = "GET"
	CommandKEYS = "KEYS"
	CommandPING = "PING"
)

type Command interface{}

type SetCommand struct {
	Key, Value []byte
}

type GetCommand struct {
	Key []byte
}

type PingCommand struct{}

type KeysCommand struct{}
