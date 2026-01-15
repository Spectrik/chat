package main

// Historie zprav v roomkach
// Ruzne typy pripojeni (websocket, nc, ...). + TLS
// Implementace databaze
//   Roomky v databazi
//   Uzivatele v databazi
//   Banlisty v databazi
// Banlisty
// Barevny text
// Format zprav v konfigu
// logovani
// Nastudovat atomic
// Predelat radne adresarovou strukturu
// Dodelat dalsi user role?
// Vyjebat context do riti kde neni potreba

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/ondrej/chat/client"
	"github.com/ondrej/chat/clientregistry"
	"github.com/ondrej/chat/commands"
	"github.com/ondrej/chat/internal/app"
	"github.com/ondrej/chat/internal/authz"
	chatconfig "github.com/ondrej/chat/internal/config"
	"github.com/ondrej/chat/internal/store/filestore"
	"github.com/ondrej/chat/room"
	"github.com/ondrej/chat/room/policies"
	"github.com/ondrej/chat/roommanager"
	"github.com/ondrej/chat/transport"
)

var log = slog.Default().With("component", "main")
func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	config, err := chatconfig.LoadConfig("config.yaml")

	if err != nil {
		log.Error("Failed to load config:", err)
	}

	// Define all the deps required for run
	store, err := filestore.NewFileStore("internal/config/userdb.txt")
	if err != nil {
		log.Error("Failed to load user database from file", err)
	}

	reg := commands.NewRegistry()
	commands.RegisterAll(reg)

	ctx := context.TODO()
	authn := filestore.NewService(store)
    authz := authz.NewService()
	cr := clientregistry.NewRegistry(store)
	rm := roommanager.NewRoomManager()

	app := app.NewApp(rm, authn, authz, cr)

	ln, err := net.Listen("tcp", config.Transport.Address)
	if err != nil {
		log.Error("Server listener start failed", err)
	}

	if config.Transport.Type == "tls" {
		cert, err := config.Certificate()
		if err != nil {
			panic("Certificate load failed")
		}

		ln = tls.NewListener(ln, &tls.Config{
			Certificates: []tls.Certificate{cert},
		})

		if err != nil {
			log.Error("TLS server listener start failed", err)
		}
	}

	defer ln.Close()

	log.Info("Chat server listening.", "address", config.Transport.Address)
	go rm.Run()

	// load rooms from config
	for _, rc := range config.Chat.Rooms {
    	pols, _ := chatconfig.BuildPolicies(rc.Policies)
    	err := app.CreateRoomNoAuth(ctx, rc.Name, pols)

		if err != nil {
			log.Error("Room failed to create!", err)
		}
	}
	policy := room.FromPolicies(policies.DefaultPolicies()).WithJoin(policies.PasswordPolicy{Password: "password123"}).Build()
	app.CreateRoomNoAuth(ctx, "testroom", &policy)
	janitorCloser := rm.StartJanitor(5 * time.Second)

	defer janitorCloser()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept:", err)
			continue
		}

		adapter := transport.NewTCPTransport(conn)
		go handleConn(adapter, app, reg)
	}
}

func handleConn(conn transport.Transport, app *app.App, reg *commands.Registry) {
	defer conn.Close()

	conn.Write("Hello there! Welcome to the chat server.")
	conn.Write("Enter your name:")

	name, _ := conn.Read()
	cr := app.CR

	for {
		if err := cr.ValidateUsername(name); err != nil {
			conn.Write(err.Error() + "\n")
			conn.Write("Enter your name:")
			name, _ = conn.Read()
			continue
		}
	break
	}

	cl := client.NewClient(conn, name)
	cr.Register(cl)

	// client writer goroutine
	go func() {
		for {
			select {
			case <-cl.Done():
				return
			case msg := <-cl.Out:
				err := conn.Write(msg)
				if err != nil {
					cl.Close()
					return
				}
			}
		}
	}()

	defer func() {
		cl.Close()
		app.Disconnect(cl)
		cr.Unregister(cl)
	}()

	cl.Send(reg.HelpText())
	for {
		line, err := conn.Read()
		if err != nil {
			break
		}

		if line == "" {
			continue
		}

		if cmd, ok := commands.ParseCommand(line); ok {
			reg.Dispatch(cl, app, cmd)

			if cl.Quit {
				cl.Send(fmt.Sprintf("Bye bye %s!", cl.ClientName()))
				cl.Close()
				return
			}
			continue
		}
	}
}
