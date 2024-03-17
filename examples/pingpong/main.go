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
	api, err := regolt.NewAPI(token, nil)
	if err != nil {
		panic(err)
	}
	socket, err := regolt.NewSocket(token.Token, nil)
	if err != nil {
		panic(err)
	}
	socket.OnAuthenticated(func(a *regolt.Authenticated) {
		fmt.Println("Authenticated.")
	})
	socket.OnMessage(func(m *regolt.Message) {
		if m.Content == "!ping" {
			api.SendMessage(m.Channel, &regolt.SendMessage{
				Content: "Pong!",
				Replies: []regolt.Reply{
					{
						ID: m.ID,
					},
				},
			})
		}
	})
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
