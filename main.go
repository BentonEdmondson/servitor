package main

import (
	"os"
	"golang.org/x/term"
	"strings"
	"mimicry/ui"
	"time"
)

// TODO: clean up most panics

func main() {
	if len(os.Args) != 2 { 
		panic("must provide 2 arguments")
	}
	oldTerminal, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil { panic(err) }
	defer term.Restore(int(os.Stdin.Fd()), oldTerminal)
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil { panic(err) }
	printRaw("")

	state := ui.Start(os.Args[1], printRaw)
	state.SetWidthHeight(width, height)

	go func() {
		for {
			time.Sleep(1 * time.Second)
			width, height, err := term.GetSize(int(os.Stdin.Fd()))
			if err != nil { panic(err) }
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
	if err != nil { panic(err) }
}
