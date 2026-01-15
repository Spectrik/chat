package commands

import (
	"strings"

	"github.com/ondrej/chat/client"
	"github.com/ondrej/chat/internal/app"
)

type Registry struct {
	cmds map[string]CmdDef
}

func NewRegistry() *Registry {
		return &Registry{
		cmds: make(map[string]CmdDef),
	}
}

func (r *Registry) Register(name string, def CmdDef) {
	r.cmds[strings.ToLower(name)] = def
}

func (r *Registry) HelpText() string {
	var b strings.Builder
	b.WriteString("Available commands:\n\n")

	for _, def := range r.cmds {
		if def.Usage == "" {
			continue
		}
		b.WriteString(def.Usage)
		b.WriteByte('\n')
	}

	return b.String()
}

func (r *Registry) Dispatch(c *client.Client, app *app.App, cmd Command) {
	def, ok := r.cmds[cmd.Name]
	if !ok {
		c.Send("Unknown command: " + cmd.Name)
		return
	}

	if def.MinArgs > 0 && len(cmd.Args) < def.MinArgs {
		if def.Usage != "" {
			c.Send("Usage: " + def.Usage)
		} else {
			c.Send("Invalid command usage")
		}
		return
	}

	if def.RequireAuth && !c.Authenticated() {
		c.Send("Not authenticated!")
		return
	}

	if err := def.Handler(c, app, cmd); err != nil {
		c.Send("ERROR: " + err.Error())
	}
}
