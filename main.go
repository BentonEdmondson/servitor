package main

import (
	"mimicry/config"
	"mimicry/ui"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

func main() {
	if len(os.Args) < 3 {
		help()
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

	go func() {
		/* Perhaps a bad design, but this holds its lock indefinitely on error, allowing for cleanup */
		err = state.Subcommand(os.Args[1], os.Args[2])
		if err != nil {
			term.Restore(int(os.Stdin.Fd()), oldTerminal)
			help()
		}
	}()

	buffer := make([]byte, 1)
	for {
		os.Stdin.Read(buffer)
		input := buffer[0]

		if input == 3 /*(ctrl+c)*/ {
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

const version = "0.0.0"

func help() {
	os.Stdout.WriteString(
"Servitor v" + version + `

Commands:
servitor open <url or @>
servitor feed <feed name>

Keybindings:
  Navigation:
  j - move down
  k - move up
  space - select the highlighted item
  c - view the creator of the highlighted item
  r - view the recipient of the highlighted item (e.g. the group it was posted to)
  a - view the actor of the activity (e.g. view the retweeter of a retweet)
  h - move back in your browser history
  l - move forward in your browser history
  g - move to the expanded item (i.e. move to the current OP)
  ctrl+c - exit the program

  Media:
  p - open the highlighted user's profile picture
  b - open the highlighted user's banner
  o - open the content of a post itself (e.g. open the video associated with a video post)
  number keys - open a link within the highlighted text

  Commands:
  :open <url or @>
  :feed <feed name>
`)
	os.Exit(0)
}
