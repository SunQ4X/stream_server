package webserver

import (
	"logger"
	"net"
	"net/http"
	"strings"
)

type WebServer struct {
	tcpListener     net.Listener
	server          *http.Server
	routerMap       map[string](map[string]http.Handler)
	notFoundHandler http.Handler
}

func NewWebServer(addr string) (*WebServer, error) {
	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	server := &WebServer{
		tcpListener: tcpListener,
		server:      &http.Server{},
		routerMap:   make(map[string](map[string]http.Handler)),
	}

	server.server.Handler = server

	return server, nil
}

func (s *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rcv := recover(); rcv != nil {
			logger.Error("panic:", rcv, r.RequestURI)
		}
	}()

	methodUpper := strings.ToUpper(r.Method)
	pathMap, ok := s.routerMap[methodUpper]
	if !ok {
		if s.notFoundHandler != nil {
			s.notFoundHandler.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
		return
	}

	handler, ok := pathMap[r.URL.Path]
	if !ok {
		if s.notFoundHandler != nil {
			s.notFoundHandler.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
		return
	}

	handler.ServeHTTP(w, r)
}

func (s *WebServer) Handle(method, path string, handler http.HandlerFunc) {
	methodUpper := strings.ToUpper(method)
	_, ok := s.routerMap[methodUpper]
	if !ok {
		s.routerMap[methodUpper] = make(map[string]http.Handler)
	}

	s.routerMap[methodUpper][path] = handler
}

func (s *WebServer) NotFoundHandler(handler http.HandlerFunc) {
	s.notFoundHandler = handler
}

func (s *WebServer) Run() {
	logger.Info("HTTP接口服务启动")
	//	if err := s.server.ListenAndServeTLS("./ssl/self.crt", "./ssl/self.key"); err != nil {
	//		fmt.Println("ListenAndServe error:", err)
	//	}
	s.server.Serve(s.tcpListener)
}
