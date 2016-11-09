package stream_server

import (
	"sync"
	"sync/atomic"

	"github.com/stream_server/rtsp"
)

type StreamServer struct {
	opts atomic.Value
	wrap sync.WaitGroup
}

func NewStreamServer(opts *Options) *StreamServer {
	server := &StreamServer{}

	server.opts.Store(opts)

	return server
}

func (s *StreamServer) getOpts() *Options {
	return s.opts.Load().(*Options)
}

func (s *StreamServer) Wrap(cb func()) {
	s.wrap.Add(1)
	go func() {
		cb()
		s.wrap.Done()
	}()
}

func (s *StreamServer) Run() {
	rtspControler := &RtspControler{s}
	rtspServer, err := rtsp.NewServer(s.getOpts().RTSPAddress, rtspControler)
	if err != nil {
		return
	}

	s.Wrap(func() {
		rtspServer.Run()
	})

	s.wrap.Wait()
}
