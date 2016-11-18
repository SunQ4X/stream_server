package web_admin

import (
	"net/http"
	"strings"
)

type Router struct {
	routerMap    map[string](map[string]http.Handler)
	PanicHandler http.Handler
}

func NewRouter() *Router {
	router := &Router{
		routerMap: make(map[string](map[string]http.Handler)),
	}

	return router
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if router.PanicHandler != nil {
		defer func() {
			if rcv := recover(); rcv != nil {
				router.PanicHandler.ServeHTTP(w, r)
			}
		}()
	}

	methodUpper = strings.ToUpper(r.method)
	pathMap, ok := router.routerMap[methodUpper]
	if !ok {
		http.NotFound(w, r)
	}

	handler, ok := pathMap[r.URL.Path]
	if !ok {
		http.NotFound(w, r)
	}

	handler.ServeHTTP(w, r)
}

func (router *Router) Handle(method, path string, handler http.HandlerFunc) {
	methodUpper = strings.ToUpper(method)
	_, ok := router.routerMap[methodUpper]
	if !ok {
		router.routerMap[methodUpper] = make(map[string]http.Handler)
	}

	router.routerMap[methodUpper][path] = handler
}
