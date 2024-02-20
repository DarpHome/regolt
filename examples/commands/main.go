package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/DarpHome/regolt"
	"github.com/DarpHome/regolt/commands"
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
	autumn, err := regolt.NewAutumnAPI(token, nil)
	if err != nil {
		panic(err)
	}
	socket, err := regolt.NewSocket(token.Token, nil)
	if err != nil {
		panic(err)
	}
	u, err := api.FetchSelf()
	if err != nil {
		panic(err)
	}
	socket.Me = u
	plugin := commands.New(api, socket, commands.Config{
		Commands: []*commands.Command{
			{
				Name: "ping",
				Callback: func(ctx *commands.Context) {
					ctx.Respond(&regolt.SendMessage{Content: "Pong!"})
				},
			},
			{
				Name: "calc",
				Options: []commands.Option{
					commands.Greedy{Option: commands.UnsignedIntOption{Name: "numbers"}},
				},
				Callback: func(ctx *commands.Context) {
					numbers := ctx.Options["numbers"].([]any)
					var result uint64
					for _, number := range numbers {
						result += number.(uint64)
					}
					ctx.Respond(&regolt.SendMessage{Content: fmt.Sprint(result)})
				},
			},
			{
				Name: "about",
				Options: []commands.Option{
					commands.StringOption{Name: "parameter", Required: false},
				},
				Callback: func(ctx *commands.Context) {
					parameter := ctx.String("parameter")
					content := ""
					if parameter != "" {
						content = fmt.Sprintf(("Running on [Regolt](https://github.com/DarpHome/regolt) %s.\n" +
							"**You passed _%s_ parameter.**"),
							regolt.Version, parameter)
					} else {
						content = fmt.Sprintf("Running on [Regolt](https://github.com/DarpHome/regolt) %s.", regolt.Version)
					}
					ctx.Respond(&regolt.SendMessage{Content: content})
				},
			},
			{
				Name: "get-hello",
				Callback: func(ctx *commands.Context) {
					hello := `package main

import "fmt"

func main() {
	fmt.Println("Hello world")
}`
					id, err := autumn.Upload("attachments", "hello.go", "", []byte(hello))
					if err != nil {
						panic(err)
					}
					ctx.Respond(&regolt.SendMessage{Attachments: []string{id}})
				},
			},
			{
				Name: "read",
				Callback: func(ctx *commands.Context) {
					if len(ctx.Message.Attachments) == 0 {
						ctx.Respond(&regolt.SendMessage{Content: "No files were found on your message."})
						return
					}
					attachment := ctx.Message.Attachments[0]
					if attachment.Size > 10000 {
						ctx.Respond(&regolt.SendMessage{Content: "Too big file. Please send less than 10 KB."})
						return
					}
					b, err := autumn.Get(attachment.Tag, attachment.ID)
					if err != nil {
						panic(err)
					}
					if len(b) > 200 {
						b = b[:200]
					}
					ctx.Respond(&regolt.SendMessage{
						Content: fmt.Sprintf("First 200 bytes: ```\n%s\n```", string(b)),
					})
				},
			},
		},
		Prefix: "!",
	})
	plugin.Install()
	// uncomment following line if you want make it work only for you
	// plugin.InstallSelfBot()
	if err = socket.Open(); err != nil {
		panic(err)
	}
	sc := make(chan os.Signal, 1)

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	if err = socket.Close(); err != nil {
		panic(err)
	}
}
