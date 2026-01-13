package main

// Heslo muzou mit jen permanent roomky
// Banlist maj jenom permanent roomky
// Jak jednoznacne identifikovat clienta?
// IP basic
// Registrace + login (moc prace na to ted) + anonymni pristup konfigurace
// RBAC (moc prace na to ted)
// Logovani
// Ruzne typy pripojeni (websocket, nc, ...). + TLS
// Databaze
// Roomky v databazi
// User permissions
// Uzivatele + hesla
// Banlisty
// Konfigurace z yamlu

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	chat "github.com/ondrej/chat/chat"
	chatconfig "github.com/ondrej/chat/config"
	transport "github.com/ondrej/chat/transport"
)

func main() {
	config, err := chatconfig.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	ln, err := net.Listen(config.Transport.Type, config.Transport.Port)
	if err != nil {
		log.Fatal(err)
	}

	if config.Transport.Type == "tls" {
		cert, err := config.Certificate()
		if err != nil {
			log.Fatal(err)
		}

		ln = tls.NewListener(ln, &tls.Config{
			Certificates: []tls.Certificate{cert},
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	defer ln.Close()

	log.Println("Chat server listening on", config.Transport.Port)


	rm := chat.NewRoomManager()
	rm.LoadRoomsFromConfig(config.Chat)
	go rm.Run()
	janitorCloser := rm.StartJanitor(5 * time.Second)

	defer janitorCloser()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept:", err)
			continue
		}

		adapter := transport.NewTCPTransport(conn)
		go handleConn(adapter, rm)
	}
}

func handleConn(conn transport.Transport, rm *chat.RoomManager) {
	defer conn.Close()

	conn.Write("Hello there! Welcome to the chat server.")
	conn.Write("Enter your name:")

	name, _ := conn.Read()
	name = trimNL(name)

	if name == "" {
		name = conn.RemoteAddr()
	}

	client := chat.NewClient(conn, name)

	// writer goroutine
	done := make(chan struct{})
	go func() {
		for msg := range client.Out {
			conn.Write(msg)
		}
		close(done)
	}()

	defer func() {
		rm.Disconnect(client)
		close(client.Out)
		<-done
	}()

	chat.HelpMessage(client.Out)

	for {
		line, err := conn.Read()
		if err != nil {
			break // disconnect
		}

		if line == "" {
			continue
		}

		if cmd, ok := chat.ParseCommand(line); ok {
			handler, exists := chat.Commands[cmd.Name]
			if !exists {
				client.Out <- "Unknown command. Try /help"
				continue
			}

			handler(client, rm, cmd)
			if client.Quit {
				client.Out <- fmt.Sprintf("Bye bye %s!", client.Name)
				return
			}
			continue
		}
	}
}

func trimNL(s string) string {
	for len(s) > 0 {
		last := s[len(s)-1]
		if last == '\n' || last == '\r' {
			s = s[:len(s)-1]
			continue
		}
		break
	}

	return s
}
