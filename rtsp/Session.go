package rtsp

import (
	"fmt"
	"net"
	"os"
)

type Session struct {
	conn net.Conn
}

func NewSession(conn net.Conn) *Session {
	sess := &Session{conn}

	return sess
}

func (sess *Session) Handle() {
	fmt.Println("Session:%s handling.", sess.conn.RemoteAddr())

	for {
		req, err := ReadRequest(sess.conn)
		if err != nil {
			break
		}

	}

	fmt.Println("Session:%s closed.", sess.conn.RemoteAddr())
}
