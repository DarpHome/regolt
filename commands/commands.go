package commands

import (
	"strconv"
	"strings"

	"github.com/DarpHome/regolt"
)

type LightContext struct {
	Manager *Commands
	Message *regolt.Message
}

type Option interface {
	GetName() string
	GetDescription() string
	Parse(*Context) (any, error)
}

type Scanner struct {
	Manager  *Commands
	Position int
	Target   string
}

func (s *Scanner) Back() {
	if s.Position == 0 {
		return
	}
	s.Position--
}

func (s *Scanner) GetByte() (byte, bool) {
	if len(s.Target) == s.Position {
		return 0, false
	}
	r := s.Target[s.Position]
	s.Position++
	return r, true
}

func (s *Scanner) Transaction(f func(s *Scanner) bool) {
	p := s.Position
	if !f(s) {
		s.Position = p
	}
}

type Context struct {
	LightContext
	Command *Command
	Prefix  string
	Scanner *Scanner
	Label   string
	Options map[string]any
}

func (ctx *Context) Integer(name string, defaultValue ...int64) int64 {
	o, ok := ctx.Options[name]
	if !ok {
		if len(defaultValue) != 0 {
			return defaultValue[0]
		}
		return 0
	}
	switch v := o.(type) {
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	}
	if len(defaultValue) != 0 {
		return defaultValue[0]
	}
	return 0
}

func (ctx *Context) String(name string, defaultValue ...string) string {
	o, ok := ctx.Options[name]
	if !ok {
		if len(defaultValue) != 0 {
			return defaultValue[0]
		}
		return ""
	}
	if v, k := o.(string); k {
		return v
	}
	if len(defaultValue) != 0 {
		return defaultValue[0]
	}
	return ""
}

type Handler interface {
	HandleCommand(*Context)
}

func (ctx *LightContext) Respond(sm *regolt.SendMessage) (*regolt.Message, error) {
	return ctx.Manager.API.SendMessage(ctx.Message.Channel, sm)
}

func (ctx *LightContext) ReactWith(s string) error {
	e := regolt.Emoji{}
	if strings.Contains(s[:1], "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		e.ID = regolt.ULID(s)
	} else {
		e.Emoji = s
	}
	return ctx.Manager.API.AddReactionToMessage(ctx.Message.Channel, ctx.Message.ID, e)
}

type GlobalCheck func(*LightContext) bool
type PrefixGetter func(*LightContext) []string
type CommandCallback func(*Context)

type Command struct {
	Name        string
	Aliases     []string
	Description string
	Options     []Option
	Callback    CommandCallback
}

type Commands struct {
	installed   bool
	Config      Config
	Handler     Handler
	Prefix      PrefixGetter
	GlobalCheck GlobalCheck
	Commands    []*Command
	API         *regolt.API
	Socket      *regolt.Socket
}

func (c *Commands) Install() *Commands {
	if !c.installed {
		c.Socket.OnMessage(c.handle)
		c.installed = true
	}
	return c
}

func (c *Commands) createLightContext(m *regolt.Message) *LightContext {
	return &LightContext{
		Manager: c,
		Message: m,
	}
}

type SignedIntOption struct {
	Name        string
	Description string
	Base        int
	BitSize     int
	Required    bool
}

func (o SignedIntOption) GetName() string {
	return o.Name
}

func (o SignedIntOption) GetDescription() string {
	return o.Description
}

func (sio SignedIntOption) Parse(ctx *Context) (any, error) {
	r := []byte{}
	for {
		b, ok := ctx.Scanner.GetByte()
		if !ok || b == ' ' {
			break
		}
		r = append(r, b)
	}
	if len(r) == 0 && sio.Required {
		return "", OptionRequired{Name: sio.Name}
	}
	base := sio.Base
	if base == 0 {
		base = 10
	}
	bitSize := sio.BitSize
	if bitSize == 0 {
		bitSize = 64
	}
	i, err := strconv.ParseInt(string(r), base, bitSize)
	return int64(i), err
}

type Greedy struct {
	Option Option
}

func (g Greedy) GetName() string {
	return g.Option.GetName()
}

func (g Greedy) GetDescription() string {
	return g.Option.GetDescription()
}

func (g Greedy) Parse(ctx *Context) (any, error) {
	a := []any{}
	for {
		b, err := g.Option.Parse(ctx)
		if err != nil {
			break
		}
		a = append(a, b)
	}
	return a, nil
}

type UnsignedIntOption struct {
	Name        string
	Description string
	Base        int
	BitSize     int
	Required    bool
}

func (o UnsignedIntOption) GetName() string {
	return o.Name
}

func (o UnsignedIntOption) GetDescription() string {
	return o.Description
}

func (uio UnsignedIntOption) Parse(ctx *Context) (any, error) {
	r := []byte{}
	for {
		b, ok := ctx.Scanner.GetByte()
		if !ok || b == ' ' {
			break
		}
		r = append(r, b)
	}
	if len(r) == 0 && uio.Required {
		return "", OptionRequired{Name: uio.Name}
	}
	base := uio.Base
	if base == 0 {
		base = 10
	}
	bitSize := uio.BitSize
	if bitSize == 0 {
		bitSize = 64
	}
	i, err := strconv.ParseUint(string(r), base, bitSize)
	return uint64(i), err
}

type StringOption struct {
	Name             string
	Description      string
	DisallowNewlines bool
	Raw              bool
	Required         bool
}

func (o StringOption) GetName() string {
	return o.Name
}

func (o StringOption) GetDescription() string {
	return o.Description
}

func (so StringOption) Parse(ctx *Context) (any, error) {
	r := []byte{}
	if so.Raw {
		for {
			b, ok := ctx.Scanner.GetByte()
			if !ok {
				break
			}
			r = append(r, b)
		}
	} else {
		quoted := false
	process:
		for i := 0; ; i++ {
			b, ok := ctx.Scanner.GetByte()
			if !ok {
				break
			}
			switch {
			case b == '\\' && quoted:
				b, ok = ctx.Scanner.GetByte()
				if !ok {
					break process
				}
				switch b {
				case '\\':
				case '\n':
					if so.DisallowNewlines {
						return "", DisallowedEscape{Which: "newlines"}
					}
				}
			case b == ' ' && !quoted:
				break process
			case (b == '"' || b == '\''):
				quoted = !quoted
				continue process
			}
			r = append(r, b)
		}
	}
	if len(r) == 0 && so.Required {
		return "", OptionRequired{Name: so.Name}
	}
	return string(r), nil
}

func (c *Commands) handle(m regolt.Message) {
	if c.Prefix == nil {
		return
	}
	pm := &m
	lctx := c.createLightContext(pm)
	if c.GlobalCheck != nil && !c.GlobalCheck(lctx) {
		return
	}
	var name, args, prefix string
	for _, p := range c.Prefix(lctx) {
		if strings.HasPrefix(m.Content, p) {
			pl := len(p)
			name = m.Content[pl:]
			i := strings.IndexRune(name, ' ')
			if i != -1 {
				name, args = name[:i], name[i+1:]
			}
			prefix = p
			break
		}
	}
	if len(name) == 0 {
		return
	}
	ctx := &Context{
		LightContext: *lctx,
		Prefix:       prefix,
		Label:        name,
		Options:      map[string]any{},
		Scanner:      &Scanner{Target: args},
	}
	if c.Handler != nil {
		c.Handler.HandleCommand(ctx)
	} else {
		c.Handle(ctx)
	}
}

func (c *Commands) Handle(ctx *Context) {
	var co *Command
	for _, d := range c.Commands {
		if d.Name == ctx.Label {
			co = d
			break
		}
		for _, a := range d.Aliases {
			if a == ctx.Label {
				co = d
				break
			}
		}
	}
	if co == nil {
		return
	}
	ctx.Command = co
	for _, o := range co.Options {
		name := o.GetName()
		r, err := o.Parse(ctx)
		if err != nil {
			return
		}
		ctx.Options[name] = r
	}
	co.Callback(ctx)
}

func selfBot() GlobalCheck {
	return func(ctx *LightContext) bool {
		return ctx.Manager.Socket.Me.ID == ctx.Message.Author
	}
}

func (c *Commands) InstallSelfBot() error {
	u, err := c.API.FetchSelf()
	if err != nil {
		return err
	}
	c.Socket.Me = u
	c.GlobalCheck = selfBot()
	return nil
}

type Config struct {
	GlobalCheck  GlobalCheck
	Handler      Handler
	Commands     []*Command
	Prefixes     []string
	Prefix       string
	PrefixGetter PrefixGetter
}

func New(api *regolt.API, socket *regolt.Socket, config Config) *Commands {
	if api == nil || socket == nil {
		return nil
	}
	var p PrefixGetter
	if config.PrefixGetter != nil {
		p = config.PrefixGetter
	} else if len(config.Prefixes) > 0 {
		p = func(ctx *LightContext) []string {
			return ctx.Manager.Config.Prefixes
		}
	} else if len(config.Prefix) > 0 {
		p = func(ctx *LightContext) []string {
			return []string{ctx.Manager.Config.Prefix}
		}
	}
	c := &Commands{
		API:      api,
		Socket:   socket,
		Commands: config.Commands,
		Config:   config,
		Handler:  config.Handler,
		Prefix:   p,
	}
	return c
}
