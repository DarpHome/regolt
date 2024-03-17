package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/DarpHome/regolt"
)

var (
	token *regolt.Token
)

func init() {
	f, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	c := struct {
		Token string `json:"token"`
	}{}
	if err = json.Unmarshal(b, &c); err != nil {
		panic(err)
	}
	token = regolt.NewUserToken(c.Token)
}

func main() {
	socket, err := regolt.NewSocket(token.Token, &regolt.SocketConfig{Cache: &regolt.GenericCache{
		Messages: &regolt.Cache2[regolt.OptimizedMessage]{
			TotalMaxSize: 1000,
		},
	}})
	if err != nil {
		panic(err)
	}
	socket.OnAuthenticated(func(a *regolt.Authenticated) {
		fmt.Println("Authenticated.")
	})
	socket.OnMessageDelete(func(event *regolt.MessageDelete) {
		message := socket.Cache.Messages.Get(event.Channel, event.MessageID)
		// if `message` is nil, then message was sent before bot started
		if message != nil {
			fmt.Printf("Message was deleted: %s, sent by %s\n", message.Content, message.Author)
		}
	})
	socket.OnMessageUpdate(func(event *regolt.MessageUpdate) {
		message := socket.Cache.Messages.Get(event.Channel, event.MessageID)
		if message != nil {
			fmt.Printf("Message %s changed; content was %s\n", message.ID, message.Content)
		}
	})
	fmt.Println("opening...")
	if err = socket.Open(); err != nil {
		panic(err)
	}
	fmt.Println("Bot is running now.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	if err = socket.Close(); err != nil {
		panic(err)
	}
}
