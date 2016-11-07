package rtsp

import (
	"fmt"
	"net"
	"os"
)

type Server struct {
	tcpListenr net.Listener
}

func NewServer(address string) *Server {
	server := &Server{}

	tcpListener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("ERROR: listen (%s) failed - %s", address, err)
		return nil, err
	}

	server.tcpListenr = tcpListener

	go func(listener net.Listener) {
		fmt.Println("Listen on %s", listener.Addr())

		for {
			clientConn, err := listener.Accept()
			if err != nil {
				//若是暂时性错误，则继续监听，否则直接退出
				if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
					fmt.Println("NOTICE: temporary Accept() failure - %s", err)
					runtime.Gosched()
					continue
				}

				if !strings.Contains(err.Error(), "use of closed network connection") {
					fmt.Println("ERROR: listener.Accept() - %s", err)
				}
				break
			}

			session := NewSession(clientConn)

			go session.Handle()
		}
	}(server.tcpListenr)

	return server
}
