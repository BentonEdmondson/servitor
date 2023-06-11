package main

import (
	"mimicry/config"
	"mimicry/ui"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// TODO: clean up most panics

func main() {
	oldTerminal, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldTerminal)
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	config, err := config.Parse()
	if err != nil {
		panic(err)
	}
	state := ui.NewState(config, width, height, printRaw)

	if len(os.Args) < 2 {
		help()
		return
	}

	switch os.Args[1] {
	case "open":
		if len(os.Args) == 3 {
			state.Open(os.Args[2])
		} else {
			help()
			return
		}
	case "feed":
		if len(os.Args) == 2 {
			state.Feed("default")
		} else if len(os.Args) == 3 {
			state.Feed(os.Args[2])
		} else {
			help()
			return
		}
	default:
		panic("expected a command as the first argument")
	}

	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			width, height, err := term.GetSize(int(os.Stdin.Fd()))
			if err != nil {
				panic(err)
			}
			state.SetWidthHeight(width, height)
		}
	}()

	buffer := make([]byte, 1)
	for {
		os.Stdin.Read(buffer)
		input := buffer[0]

		if input == 3 /*(ctrl+c)*/ || input == 'q' {
			printRaw("")
			return
		}

		state.Update(input)
	}
}

func printRaw(output string) {
	output = strings.ReplaceAll(output, "\n", "\r\n")
	_, err := os.Stdout.WriteString("\x1b[0;0H\x1b[2J" + output)
	if err != nil {
		panic(err)
	}
}

func help() {
	os.Stdout.WriteString("here's the help page\n")
}
