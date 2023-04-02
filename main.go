package main

import (
	"fmt"
	"mimicry/kinds"
	"os"
	"encoding/json"
)

// TODO: when returning errors, use zero value for return
// also change all error messages to using sprintf-style
// formatting, all lowercase, and no punctuation

// TODO: get rid of Raw, just use jtp.Get and then stringify the result

func main() {

	link := os.Args[len(os.Args)-1]
	command := os.Args[1]

	item, err := kinds.FetchUserInput(link)
	if err != nil {
		panic(err)
	}

	if command == "raw" {
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(item); err != nil {
			panic(err)
		}
		return
	}

	// if narrowed, ok := content.(kinds.Post); ok {
	// 	if str, err := narrowed.Preview(); err != nil {
	// 		panic(err)
	// 	} else {
	// 		fmt.Print(str)
	// 	}
	// 	return
	// }

	if str, err := item.String(90); err != nil {
		panic(err)
	} else {
		fmt.Print(str)
	}
}
