package chat

import (
	"strings"
)

type Command struct {
	Name string
	Args []string
	Raw  string
}

type HandlerFunc func(c *Client, rm *RoomManager, cmd Command)

var Commands = map[string]HandlerFunc{
		"help": helpCmd,
		"join": joinCmd,
		"leave": leaveCmd,
		"quit": quitCmd,
		"roomlist": roomlistCmd,
		"say": sayCmd,
}

func ParseCommand(line string) (Command, bool) {
	if len(line) == 0 || line[0] != '/' {
		return Command{}, false
	}

	fields := strings.Fields(line[1:])
	if len(fields) == 0 {
		return Command{Name: ""}, true
	}

	return Command{
		Name: strings.ToLower(fields[0]),
		Args: fields[1:],
		Raw:  line,
	}, true
}

func helpCmd(c *Client, rm *RoomManager, cmd Command) {
	HelpMessage(c.Out)
}

func sayCmd(c *Client, rm *RoomManager, cmd Command) {
	if len(cmd.Args) < 2 {
		c.Out <- "Usage: /say <roomname> <message>"
		return
	}

	roomname := cmd.Args[0]
	message := strings.Join(cmd.Args[1:], " ")
	rm.do(func() {
		if _, ok := rm.memberOf[c][roomname]; !ok {
			c.Out <- "You are not in room: " + roomname
			return
		}

		r := rm.rooms[roomname]
		if r == nil {
			c.Out <- "No such room: " + roomname
			return
		}
		r.Say(c, message)
	})
}

func quitCmd(c *Client, rm *RoomManager, cmd Command) {
	c.Quit = true
}

func roomlistCmd(c *Client, rm *RoomManager, cmd Command) {
	if rm.memberOf[c] == nil {
		c.Out <- "You are not in any rooms."
		return
	}

	c.Out <- "You are in the following rooms:"
	for name := range rm.memberOf[c] {
		c.Out <- name
	}
}

func joinCmd(c *Client, rm *RoomManager, cmd Command) {
	if len(cmd.Args) < 1 {
		c.Out <- "Usage: /join <roomname>"
		return
	}

	roomname := cmd.Args[0]
	rm.JoinRoom(roomname, c)
}

func leaveCmd(c *Client, rm *RoomManager, cmd Command) {
	if len(cmd.Args) < 1 {
		c.Out <- "Usage: /leave <roomname>"
		return
	}

	roomname := cmd.Args[0]
	rm.LeaveRoom(roomname, c)
}
