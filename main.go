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
	if len(os.Args) < 3 {
		help()
		return
	}

	config, err := config.Parse()
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		return
	}

	oldTerminal, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldTerminal)
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	state := ui.NewState(config, width, height, printRaw)
	err = state.Subcommand(os.Args[1], os.Args[2])
	if err != nil {
		term.Restore(int(os.Stdin.Fd()), oldTerminal)
		help()
		return
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
		go state.Update(input)
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
