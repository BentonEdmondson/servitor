package main

import (
	"encoding/json"
	"fmt"
	"mimicry/kinds"
	"os"
)

// TODO: when returning errors, use zero value for return
// also change all error messages to using sprintf-style
// formatting, all lowercase, and no punctuation

func main() {
	link := os.Args[len(os.Args)-1]
	command := os.Args[1]

	content, err := kinds.FetchUnknown(link)
	if err != nil {
		panic(err)
	}

	if command == "raw" {
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(content); err != nil {
			panic(err)
		}
		return
	}

	if str, err := content.String(); err != nil {
		panic(err)
	} else {
		fmt.Println(str)
	}
}
