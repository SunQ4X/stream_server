package web_admin

import (
	"net/http"

	"github.com/stream_server/media"
)

type HttpServer struct {
	server *http.Server
}

func NewHttpServer(address string) (*HttpServer, error) {
	httpServer := &HttpServer{
		server: &http.Server{
			Addr: address,
		},
	}

	router := &Router{}
	router.Handle("GET", "/MediaSession", httpServer.GetMediaSessions)
	router.Handle("PUT", "/MediaSession", httpServer.AddMediaSessions)
	router.Handle("POST", "/MediaSession", httpServer.UpdateMediaSessions)
	router.Handle("DELETE", "/MediaSession", httpServer.DeleteMediaSessions)

	httpServer.server.Handler = router

	return httpServer, nil
}

func (server *HttpServer) Run() {
	server.server.ListenAndServe()
}

func (server *HttpServer) GetMediaSessions(w http.ResponseWriter, r *http.Request) {

}

func (server *HttpServer) AddMediaSession(w http.ResponseWriter, r *http.Request) {

}

func (server *HttpServer) UpdateMediaSession(w http.ResponseWriter, r *http.Request) {

}

func (server *HttpServer) DeleteMediaSession(w http.ResponseWriter, r *http.Request) {

}
