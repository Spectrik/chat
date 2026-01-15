package commands

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/ondrej/chat/client"
	"github.com/ondrej/chat/internal/app"
	"github.com/ondrej/chat/internal/authn"
	"github.com/ondrej/chat/protocol"
)

var log = slog.Default().With("component", "commands")

type Command struct {
	Name string
	Args []string
	Raw  string
}

type HandlerFunc func(c *client.Client, app *app.App, cmd Command) error

type CmdDef struct {
	Usage       string
	MinArgs     int
	RequireAuth bool
	Authn		[]authn.Authenticator
	Handler     HandlerFunc
}

func RegisterAll(r *Registry) {
	r.Register("create", CmdDef{
		Usage:       "/create <roomname>",
		MinArgs:     1,
		RequireAuth: true,
		Handler:     createCmd,
	})

	r.Register("join", CmdDef{
		Usage:       "/join <roomname> [password]",
		MinArgs:     1,
		RequireAuth: false,
		Handler:     joinCmd,
	})

	r.Register("role", CmdDef{
		Usage:       "/role",
		MinArgs:     0,
		RequireAuth: true,
		Handler:     userRolesCmd,
	})

	r.Register("login", CmdDef{
		Usage:       "/login <user> <password>",
		MinArgs:     2,
		RequireAuth: false,
		Handler:     loginCmd,
	})

	r.Register("msg", CmdDef{
		Usage:       "/msg <user> <message>",
		MinArgs:     2,
		RequireAuth: false,
		Handler:     directMessageCmd,
	})

	r.Register("logout", CmdDef{
		Usage:       "/logout",
		MinArgs:     0,
		RequireAuth: true,
		Handler:     logoutCmd,
	})

	r.Register("mute", CmdDef{
		Usage:       "/mute <room> <user> <duration>",
		MinArgs:     3,
		RequireAuth: true,
		Handler:     muteCmd,
	})

	r.Register("leave", CmdDef{
		Usage:       "/leave <roomname>",
		MinArgs:     1,
		RequireAuth: false,
		Handler:     leaveCmd,
	})

	r.Register("say", CmdDef{
		Usage:       "/say <roomname> <message...>",
		MinArgs:     2,
		RequireAuth: false,
		Handler:     sayCmd,
	})

	r.Register("roomlist", CmdDef{
		Usage:       "/roomlist",
		MinArgs:     0,
		RequireAuth: false,
		Handler:     roomlistCmd,
	})

	r.Register("quit", CmdDef{
		Usage:       "/quit",
		MinArgs:     0,
		RequireAuth: false,
		Handler:     quitCmd,
	})

	r.Register("help", CmdDef{
		Usage:       "/help",
		MinArgs:     0,
		RequireAuth: false,
		Handler: func(c *client.Client, app *app.App, cmd Command) error {
			c.Send(r.HelpText())
			return nil
		},
	})
}

func logoutCmd(c *client.Client, app *app.App, cmd Command) error {
	if err := app.Disconnect(c); err != nil {
		return err
	}

	c.Identity.User = nil
	c.Send("You've been logged out.")
	return nil
}

func muteCmd(c *client.Client, app *app.App, cmd Command) error {
	r := cmd.Args[0]
	u := cmd.Args[1]
	d, err := strconv.Atoi(cmd.Args[2])

	if err != nil {
		return fmt.Errorf("Error in muteCmd handler: %s", err)
    }

	ctx := context.TODO()
	return app.Mute(ctx, protocol.MuteArgs{
		Room: r,
		Actor: c,
		Target: u,
		Duration: d,
	})
}

func createCmd(c *client.Client, app *app.App, cmd Command) error {
	roomname := strings.TrimSpace(cmd.Args[0])
	if roomname == "" {
		return fmt.Errorf("room name cannot be empty")
	}

	ctx := context.TODO()
	if err := app.CreateRoom(ctx, c ,roomname, nil); err != nil {
		return err
	}

	c.Send("Room created: " + roomname)
	return nil
}

func directMessageCmd(c *client.Client, app *app.App, cmd Command) error {
	username := cmd.Args[0]
	message := strings.Join(cmd.Args[1:], " ")

	return app.DirectMessage(protocol.DirectMessageArgs{
		Sayer: c,
		User: username,
		Text: message,
	})
}

func loginCmd(c *client.Client, app *app.App, cmd Command) error {
	username := cmd.Args[0]
	password := cmd.Args[1]

	return app.Login(protocol.LoginArgs{
		Client: c,
		Username: username,
		Password: password,
	})
}

func userRolesCmd(c *client.Client, app *app.App, cmd Command) error {
	c.Send("You have the following role: \n\n" + c.Identity.User.Role.String())
	return nil
}

func sayCmd(c *client.Client, app *app.App, cmd Command) error {
	roomname := cmd.Args[0]
	message := strings.Join(cmd.Args[1:], " ")
	ctx := context.TODO()

	return app.Say(ctx, protocol.SayArgs{
		Sayer: c,
		Room: roomname,
		Text:  message,
	})
}

func quitCmd(c *client.Client, app *app.App, cmd Command) error {
	c.Quit = true
	return nil
}

func roomlistCmd(c *client.Client, app *app.App, cmd Command) error {
	return app.ClientRoomList(c)
}

func joinCmd(c *client.Client, app *app.App, cmd Command) error {
	var password = ""
	roomname := cmd.Args[0]
	ctx := context.TODO()

	if len(cmd.Args) >= 2 {
		password = cmd.Args[1]
	}

	return app.JoinRoom(ctx, protocol.JoinRoomArgs{
		Password: password,
		Client: c,
		Room: roomname,
	})
}

func leaveCmd(c *client.Client, app *app.App, cmd Command) error {
	ctx := context.TODO()
	roomname := cmd.Args[0]
	if err := app.LeaveRoom(ctx, c, roomname); err != nil {
		return  err
	}

	return nil
}
