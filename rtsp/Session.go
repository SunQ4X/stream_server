package rtsp

import (
	"fmt"
	"net"
)

type Session struct {
	conn net.Conn
}

func NewSession(conn net.Conn) *Session {
	sess := &Session{conn: conn}

	return sess
}

func (sess *Session) Handle(controler Controler) {
	fmt.Println("------ Session[", sess.conn.RemoteAddr(), "] : handling ------")

	for {
		req, err := ReadRequest(sess.conn)
		if err != nil {
			break
		}

		fmt.Println("------ Session[", sess.conn.RemoteAddr(), "] : get request ------ \n", req)
		//TODO
		//处理RTSP请求
		resp := controler.Control(req)

		sess.conn.Write([]byte(resp.String()))
		fmt.Println("------ Session[", sess.conn.RemoteAddr(), "] : set response ------ \n", resp)
	}

	sess.conn.Close()
	fmt.Println("------ Session[", sess.conn.RemoteAddr(), "] : closed ------")
}
