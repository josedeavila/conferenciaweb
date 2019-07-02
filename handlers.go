package main

import (
	"fmt"
	"log"
	"time"

	r "github.com/dancannon/gorethink"
	"github.com/mitchellh/mapstructure"
)

const (
	ChannelStop = iota
	UserStop
	MessageStop
)

type Message struct {
	ChannelId string      `json:"channelId"`
	UserId    string      `json:"userId"`
	Data      interface{} `json:"data"`
	Acao      string      `json:"acao"`
}

type Channel struct {
	Id   string `json:"id" gorethink:"id,omitempty"`
	Nome string `json:"nome" gorethink:"name"`
	Acao string `json:"acao"`
}

type User struct {
	Id        string `gorethink:"id,omitempty"`
	Nome      string `gorethink:"nome"`
	ChannelId string `gorethink:"channelId"`
	Acao      string `json:"acao"`
}

type ChannelMessage struct {
	Id        string    `gorethink:"id,omitempty"`
	ChannelId string    `gorethink:"channelId"`
	UserId    string    `gorethink:"userId"`
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
	client.UserId = user.Id
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
	var channelMessage ChannelMessage
	err := mapstructure.Decode(data, &channelMessage)
	if err != nil {
		client.send <- Message{"error", "error", err.Error(), "addChannelMessage 1 iderror"}
	}
	go func() {
		channelMessage.CreatedAt = time.Now()
		channelMessage.UserId = client.UserId
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
	var channelM ChannelMessage
	if err := mapstructure.Decode(data, &channelM); err != nil {
		log.Fatal(err)
	}

	fmt.Println("subscribeChannelMessage 1 : ", channelM)

	go func() {

		if channelM.ChannelId == "" {
			return
		}

		stop := client.NewStopChannel(MessageStop)
		cursor, err := r.Table("message").
			OrderBy(r.OrderByOpts{Index: r.Desc("createdAt")}).
			Filter(r.Row.Field("channelId").Eq(channelM.ChannelId)).
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)

		if err != nil {
			client.send <- Message{"error", "error", err.Error(), "iderror"}
			return
		}
		changeFeedHelper(cursor, "message", client.send, stop)
	}()
}

func unsubscribeChannelMessage(client *Client, data interface{}) {
	client.StopForKey(MessageStop)
}

func addChannel(client *Client, data interface{}) {
	var channel Channel
	err := mapstructure.Decode(data, &channel)
	if err != nil {
		client.send <- Message{"error", "error", err.Error(), "iderror"}
		return
	}
	go func() {
		err = r.Table("channel").
			Insert(channel).
			Exec(client.session)
		if err != nil {
			client.send <- Message{"error", "error", err.Error(), "iderror"}
		}
	}()
}

func subscribeChannel(client *Client, data interface{}) {
	go func() {
		stop := client.NewStopChannel(ChannelStop)
		cursor, err := r.Table("channel").
			Changes(r.ChangesOpts{IncludeInitial: true}).
			Run(client.session)
		if err != nil {
			client.send <- Message{"error", "error", err.Error(), "iderror"}
			return
		}
		changeFeedHelper(cursor, "channel", client.send, stop)
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
