package rtsp_server

import (
	"logger"
	"net"
	"runtime"
	"utility"
)

type RtspServer struct {
	tcpListener net.Listener
}

func NewRtspServer() (*RtspServer, error) {
	server := &RtspServer{}

	tcpListener, err := net.Listen("tcp", utility.GetOptions().RTSPAddress)
	if err != nil {
		return nil, err
	}

	server.tcpListener = tcpListener

	return server, nil
}

func (s *RtspServer) Run() {
	logger.Info("RTSP服务启动")

	for {
		clientConn, err := s.tcpListener.Accept()
		if err != nil {
			logger.Error("rtsp listen err:", err)
			//若是暂时性错误，则继续监听，否则直接退出
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				runtime.Gosched()
				continue
			}

			break
		}

		conn := NewRtspClientConnection(clientConn)

		go conn.Handle()
	}
}
