package rtsp

import (
	"fmt"
	"net"
	"runtime"
)

type Server struct {
	tcpListener net.Listener
	controler   Controler
}

func NewServer(address string, controler Controler) (*Server, error) {
	server := &Server{}

	tcpListener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("ERROR: listen (", address, ") failed -", err)
		return nil, err
	}

	server.tcpListener = tcpListener
	server.controler = controler

	return server, nil
}

func (s *Server) Run() {
	fmt.Println("RTSP Listen on", s.tcpListener.Addr())

	for {
		clientConn, err := s.tcpListener.Accept()
		if err != nil {
			//若是暂时性错误，则继续监听，否则直接退出
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				fmt.Println("NOTICE: temporary Accept() failure -", err)
				runtime.Gosched()
				continue
			}

			break
		}

		session := NewSession(clientConn)

		go session.Handle(s.controler)
	}

	fmt.Println("RTSP Stop listenning on", s.tcpListener.Addr())
}
