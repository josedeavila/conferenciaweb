package main

import (
	"fmt"
	"log"
	"time"

	r "github.com/dancannon/gorethink"
	"github.com/mitchellh/mapstructure"

	"./sronacional/modelo"

)

const (
	ChannelStop = iota
	UserStop
	MessageStop
)

type Message struct {
	IdUnidade string      `json:"idunidade"`
	IdUser    string      `json:"iduser"`
	Data      interface{} `json:"data"`
	Acao      string      `json:"acao"`
}

type User struct {
	Id        string `gorethink:"id,omitempty"`
	Nome      string `gorethink:"nome"`
	IdUnidade string `gorethink:"idunidade"`
	Acao      string `json:"acao"`
}

type UnidadeMessage struct {
	Id        string    `gorethink:"id,omitempty"`
	IdUnidade string    `gorethink:"idunidade"`
	IdUser    string    `gorethink:"iduser"`
	Data      string    `gorethink:"data"`
	CreatedAt time.Time `gorethink:"createdAt"`
}

func editUser(client *Client, data interface{}) {
	// Logando no console do Webserver
	fmt.Println("handlers editUser: ", data)
	var user User
	err := mapstructure.Decode(data, &user)
	if err != nil {
		client.send <- Message{"error", "error", err.Error(), "iderror"}
		return
	}
	fmt.Println("handlers editUser user : ", user)
	client.IdUser = user.Id
	go func() {
		_, err := r.Table("user").
			Get(client.id).
			Update(user).
			RunWrite(client.session)
		if err != nil {
			client.send <- Message{"error", "error", err.Error(), "iderror"}
		}
	}()
}

func subscribeUser(client *Client, data interface{}) {
	fmt.Println("subscribeUser : ", data)
	go func() {
		stop := client.NewStopChannel(UserStop)
		cursor, err := r.Table("user").
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)

		if err != nil {
			client.send <- Message{"error", "error", err.Error(), "iderror"}
			return
		}
		changeFeedHelper(cursor, "user", client.send, stop)
	}()
}

func unsubscribeUser(client *Client, data interface{}) {
	client.StopForKey(UserStop)
}

func addChannelMessage(client *Client, data interface{}) {
	fmt.Println("addChannelMessage : ", data)
	var channelMessage UnidadeMessage
	err := mapstructure.Decode(data, &channelMessage)
	if err != nil {
		client.send <- Message{"error", "error", err.Error(), "addChannelMessage 1 iderror"}
	}
	go func() {
		channelMessage.CreatedAt = time.Now()
		channelMessage.IdUser = client.IdUser
		err := r.Table("message").
			Insert(channelMessage).
			Exec(client.session)
		if err != nil {
			client.send <- Message{"error", "error", err.Error(), "addChannelMessage 2 iderror"}
		}
	}()
}

func subscribeChannelMessage(client *Client, data interface{}) {
	fmt.Println("subscribeChannelMessage : ", data)
	var channelM UnidadeMessage
	if err := mapstructure.Decode(data, &channelM); err != nil {
		log.Fatal(err)
	}

	fmt.Println("subscribeChannelMessage 1 : ", channelM)

	go func() {

		if channelM.IdUnidade == "" {
			return
		}

		stop := client.NewStopChannel(MessageStop)
		cursor, err := r.Table("message").
			OrderBy(r.OrderByOpts{Index: r.Desc("createdAt")}).
			Filter(r.Row.Field("IdUnidade").Eq(channelM.IdUnidade)).
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)

		if err != nil {
			client.send <- Message{"error", "error", err.Error(), "iderror"}
			return
		}
		changeFeedHelper(cursor, "message", client.send, stop)
	}()
}

func subscribeUnidade(client *Client, data interface{}) {
	fmt.Println("subscribeUnidade : ", data)
	var channelM Unitizador
	if err := mapstructure.Decode(data, &channelM); err != nil {
		log.Fatal(err)
	}

	fmt.Println("subscribeUnidade 1 : ", channelM)

	go func() {

		if channelM.IdUnidade == "" {
			return
		}

		stop := client.NewStopChannel(MessageStop)
		cursor, err := r.Table("unitizador").
			OrderBy(r.OrderByOpts{Index: r.Desc("datacriacao")}).
			Filter(r.Row.Field("iddestino").Eq(channelM.IdUnidadeDestino)).
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)

		if err != nil {
			client.send <- Message{"error", "error", err.Error(), "iderror"}
			return
		}
		changeFeedHelper(cursor, "unitizador", client.send, stop)
	}()
}

func unsubscribeChannelMessage(client *Client, data interface{}) {
	client.StopForKey(MessageStop)
}

func subscribeChannel(client *Client, data interface{}) {
	go func() {
		stop := client.NewStopChannel(ChannelStop)
		cursor, err := r.Table("unidade").
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)
		if err != nil {
			client.send <- Message{"error", "error", err.Error(), "iderror"}
			return
		}
		changeFeedHelper(cursor, "unidade", client.send, stop)
	}()
}

func unsubscribeChannel(client *Client, data interface{}) {
	client.StopForKey(ChannelStop)
}

func changeFeedHelper(cursor *r.Cursor, changeEventName string,
	send chan<- Message, stop <-chan bool) {
	change := make(chan r.ChangeResponse)
	cursor.Listen(change)
	for {
		eventName := ""
		var data interface{}
		select {
		case <-stop:
			cursor.Close()
			return
		case val := <-change:
			if val.NewValue != nil && val.OldValue == nil {
				eventName = changeEventName + " add"
				data = val.NewValue
			} else if val.NewValue == nil && val.OldValue != nil {
				eventName = changeEventName + " remove"
				data = val.OldValue
			} else if val.NewValue != nil && val.OldValue != nil {
				eventName = changeEventName + " edit"
				data = val.NewValue
			}
			fmt.Println("changeFeedHelper send 1 : ", data, eventName)
			var message Message

			err := mapstructure.Decode(data, &message)
			if err != nil {
				log.Fatal(err)
			}

			send <- message
		}
	}
}
