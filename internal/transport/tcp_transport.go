package transport

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/tidwall/resp"
	"github.com/vector-ops/memstore/internal/protocol"
)

type TCPTransport struct {
	conn  net.Conn
	msgCh chan Message
	delCh chan Transport
}

func (t *TCPTransport) Ping() (string, error) {
	return "", fmt.Errorf("Ping not implemented")
}

func (t *TCPTransport) GetRemoteAddress() string {
	return t.conn.RemoteAddr().String()
}

func (t *TCPTransport) GetLocalAddress() string {
	return t.conn.LocalAddr().String()
}

func (p *TCPTransport) Send(msg []byte) (int, error) {
	return p.conn.Write(msg)
}

func (p *TCPTransport) Read() ([]byte, error) {
	var msg []byte
	_, err := p.conn.Read(msg)
	return msg, err
}

func NewTCPTransport(conn net.Conn, msgCh chan Message, delCh chan Transport) Transport {
	return &TCPTransport{
		conn:  conn,
		msgCh: msgCh,
		delCh: delCh,
	}
}

func (p *TCPTransport) ReadLoop() error {
	rd := resp.NewReader(p.conn)
	for {
		v, _, err := rd.ReadValue()
		if err == io.EOF {
			p.delCh <- p
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if v.Type() == resp.Array {
			for _, c := range v.Array() {
				switch strings.ToUpper(c.String()) {
				case protocol.CommandGET:
					if len(v.Array()) != 2 {
						return fmt.Errorf("invalid number of variables for GET command")
					}
					cmd := protocol.GetCommand{
						Key: v.Array()[1].Bytes(),
					}
					p.msgCh <- Message{
						Cmd:       cmd,
						Transport: p,
					}
				case protocol.CommandSET:
					if len(v.Array()) != 3 {
						return fmt.Errorf("invalid number of variables for SET command")
					}
					cmd := protocol.SetCommand{
						Key:   v.Array()[1].Bytes(),
						Value: v.Array()[2].Bytes(),
					}
					p.msgCh <- Message{
						Cmd:       cmd,
						Transport: p,
					}
				case protocol.CommandKEYS:
					if len(v.Array()) != 1 {
						return fmt.Errorf("invalid number of variables for KEYS command")
					}
					cmd := protocol.KeysCommand{}
					p.msgCh <- Message{
						Cmd:       cmd,
						Transport: p,
					}
				case protocol.CommandPING:
					if len(v.Array()) != 1 {
						return fmt.Errorf("invalid number of variables for PING command")
					}
					cmd := protocol.PingCommand{}
					p.msgCh <- Message{
						Cmd:       cmd,
						Transport: p,
					}
				default:
				}
			}
		}
	}
	return nil
}
