package main

import (
	"encoding/json"
	"fmt"
	"log"

	r "github.com/dancannon/gorethink"
	"github.com/gorilla/websocket"
)

type FindHandler func(string) (Handler, bool)

type Client struct {
	send         chan Message
	socket       *websocket.Conn
	findHandler  FindHandler
	session      *r.Session
	stopChannels map[int]chan bool
	id           string
	IdUser       string
}

func (c *Client) NewStopChannel(stopKey int) chan bool {
	c.StopForKey(stopKey)
	stop := make(chan bool)
	c.stopChannels[stopKey] = stop
	return stop
}

func (c *Client) StopForKey(key int) {
	if ch, found := c.stopChannels[key]; found {
		ch <- true
		delete(c.stopChannels, key)
	}
}

func (client *Client) Read() {
	var message Message
	for {
		if err := client.socket.ReadJSON(&message); err != nil {
			break
		}

		fmt.Println("client Read : ", message)
		//if handler, found := client.findHandler(message.UserId); found {
		if handler, found := client.findHandler(message.Acao); found {
			fmt.Println("client Read 1: ", client)
			fmt.Println("client Read 1: ", message)
			fmt.Println("client Read 1: ", handler)
			fmt.Println("client Read 1: ", found)
			handler(client, message)
		}
	}
	client.socket.Close()
}

func (client *Client) Write() {
	for msg := range client.send {
		fmt.Println("client Write : ", msg)
		if err := client.socket.WriteJSON(msg); err != nil {
			fmt.Println("client Write erro: ", err)
			break
		}
	}
	client.socket.Close()
}

func (c *Client) Close() {
	for _, ch := range c.stopChannels {
		ch <- true
	}
	close(c.send)
	// delete user
	r.Table("user").Get(c.id).Delete().Exec(c.session)
}

func NewClient(socket *websocket.Conn, findHandler FindHandler,
	session *r.Session) *Client {

	// Vamos ler a mensagem recebida via Websocket
	msgType, msg, err := socket.ReadMessage()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	// Logando no console do Webserver
	fmt.Println("NewClient recebida 1: ", msgType, string(msg))

	var user User
	//user.Nome = "81324065"

	if err := json.Unmarshal([]byte(string(msg)), &user); err != nil {
		log.Fatal(err)
	}
	// Logando no console do Webserver
	fmt.Println("NewClient recebida 2: ", user)
	res, err := r.Table("user").Insert(user).RunWrite(session)
	if err != nil {
		log.Println(err.Error())
	}
	var id string
	fmt.Println("NewClient recebida 3: ", res, res.GeneratedKeys)
	if len(res.GeneratedKeys) > 0 {
		id = res.GeneratedKeys[0]
	}else{
		fmt.Println("NewClient recebida 3 else: ", res, res.GeneratedKeys)
	}

	return &Client{
		send:         make(chan Message),
		socket:       socket,
		findHandler:  findHandler,
		session:      session,
		stopChannels: make(map[int]chan bool),
		id:           id,
		IdUser:       user.Id,
	}
}
