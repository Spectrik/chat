package commands

import "strings"

func ParseCommand(line string) (Command, bool) {
	if len(line) == 0 || line[0] != '/' {
		return Command{}, false
	}

	fields := strings.Fields(line[1:])
	if len(fields) == 0 {
		return Command{}, false
	}

	return Command{
		Name: strings.ToLower(fields[0]),
		Args: fields[1:],
		Raw:  line,
	}, true
}
