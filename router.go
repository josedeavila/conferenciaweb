package main

import (
	"fmt"
	"net/http"

	r "github.com/dancannon/gorethink"
	"github.com/gorilla/websocket"
)

type Handler func(*Client, interface{})

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Router struct {
	rules   map[string]Handler
	session *r.Session
}

func NewRouter(session *r.Session) *Router {
	return &Router{
		rules:   make(map[string]Handler),
		session: session,
	}
}

func (r *Router) Handle(msgName string, handler Handler) {
	fmt.Printf("Handle recebida: ", msgName, handler)
	r.rules[msgName] = handler
}

func (r *Router) FindHandler(msgName string) (Handler, bool) {
	handler, found := r.rules[msgName]
	fmt.Printf("FindHandler recebida: ", msgName, handler)
	return handler, found
}

func (e *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Printf("ServeHTTP recebida: ", socket)

	client := NewClient(socket, e.FindHandler, e.session)
	fmt.Println("ServeHTTP NewClient : ", client)
	defer client.Close()
	go client.Write()
	client.Read()

}
