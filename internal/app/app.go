package app

import (
	"context"
	"fmt"
	"time"

	"github.com/ondrej/chat/client"
	"github.com/ondrej/chat/clientregistry"
	"github.com/ondrej/chat/internal/authn"
	"github.com/ondrej/chat/internal/authz"
	"github.com/ondrej/chat/internal/store"
	"github.com/ondrej/chat/protocol"
	"github.com/ondrej/chat/room"
	"github.com/ondrej/chat/room/policies"
	"github.com/ondrej/chat/room/roomcmd"
	"github.com/ondrej/chat/roommanager"
)

type App struct {
	rm    *roommanager.RoomManager
	CR	  *clientregistry.Registry
	users store.UserStore
	authn authn.Authenticator
	authz authz.Authorizer
}

func NewApp(rm *roommanager.RoomManager, authn authn.Authenticator, authz authz.Authorizer, cr *clientregistry.Registry) *App {
	return &App{
		rm:    rm,
		authn: authn,
		authz: authz,
		CR: cr,
	}
}

func (a *App) CreateRoom(ctx context.Context, c *client.Client, roomname string, pols *room.PolicySet) error {
	if err := a.authz.CanCreateRoom(c); err != nil {
		return err
	}

	if pols == nil {
		tmp := policies.DefaultPolicies()
		pols = &tmp
	}

	r, e := a.rm.CreateRoom(roomname, *pols)
	if e != nil {
		return e
	}

	go pumpRoomEvents(r)
	return nil
}

func (a *App) CreateRoomNoAuth(ctx context.Context, roomname string, pols *room.PolicySet) error {
	if pols == nil {
		tmp := policies.DefaultPolicies()
		pols = &tmp
	}

	r, e := a.rm.CreateRoom(roomname, *pols)
	if e != nil {
		return e
	}

	go pumpRoomEvents(r)
	return nil
}

func (a *App) JoinRoom(ctx context.Context, args protocol.JoinRoomArgs) error {
	r, ok := a.rm.GetRoom(args.Room)
	if !ok {
		return fmt.Errorf("Room %s does not exist", args.Room)
	}

	if err := r.Exec(ctx, roomcmd.JoinCmd{Args: args}); err != nil {
		return err
	}

	a.rm.DoSync(func() {
		if a.rm.MemberOf[args.Client] == nil {
			a.rm.MemberOf[args.Client] = make(map[string]struct{})
		}

		a.rm.MemberOf[args.Client][args.Room] = struct{}{}
	})

	return nil
}

func (a *App) DirectMessage(args protocol.DirectMessageArgs) error {
    msg := fmt.Sprintf("Direct message from user %s: %s", args.Sayer.ClientName(), args.Text)

    return a.CR.WithClientByName(args.User, func(c *client.Client) error {
        if ok := c.Send(msg); !ok {
            return fmt.Errorf("failed to send direct message")
        }
        return nil
    })
}

func (a *App) LeaveRoom(ctx context.Context, c *client.Client, roomname string) error {
	r, ok := a.rm.GetRoom(roomname)
	if !ok {
		return fmt.Errorf("Room %s does not exist", roomname)
	}

	if err := r.Exec(ctx, roomcmd.LeaveCmd{Client: c,}); err != nil {
		return err
	}

	a.rm.DoSync(func() {
		if m := a.rm.MemberOf[c]; m != nil {
			delete(m, roomname)
			if len(m) == 0 {
				delete(a.rm.MemberOf, c)
			}
		}
	})

	return nil
}

func (a *App) Disconnect(c *client.Client) error {
	rooms := a.rm.ListRoomsOf(c)
	for _, r := range rooms { _ = a.LeaveRoom(context.TODO(), c, r) } // TODO
	return nil
}

func (a *App) Say(ctx context.Context, args protocol.SayArgs) error {
	r, exists := a.rm.GetRoom(args.Room)
	if !exists {
		return fmt.Errorf("Room %s does not exist", args.Room)
	}

	return r.Exec(ctx, roomcmd.SayCmd{
		Client: args.Sayer,
		Text: args.Text,
	})
}

func (a *App) Mute(ctx context.Context, args protocol.MuteArgs) error {
	r, exists := a.rm.GetRoom(args.Room)
	if !exists {
		return fmt.Errorf("Room does not exist!")
	}

	if err := a.authz.CanMute(args.Actor); err != nil {
		return err
	}

	return r.Exec(ctx, roomcmd.MuteCmd{
		Actor:      args.Actor,
		Target:     args.Target,
		Duration:   time.Duration(args.Duration) * time.Minute,
	})
}

func (a *App) Login(args protocol.LoginArgs) error {
	if args.Client.Identity.User != nil {
		return fmt.Errorf("You are already logged in")
	}

	user, err := a.authn.Authenticate(args.Username, args.Password)
	if err != nil {
		return fmt.Errorf("Login failed: %s", err)
	}

	a.CR.Unregister(args.Client)
	args.Client.Identity.User = user
	a.CR.Register(args.Client)

	args.Client.Send("Login success!")
	return nil
}

func (a *App) ClientRoomList(c *client.Client) error {
	rms := a.rm.ListRoomsOf(c)

	c.Send("You are in the following rooms:")
	if len(rms) == 0 {
		c.Send("<none>")
		return nil
	}

	for _, name := range rms {
		c.Send("- " + name)
	}

	return nil
}
