package main

import (
	"encoding/json"
	"fmt"
	"mimicry/kinds"
	"net/url"
	"os"
)

func main() {
	link := os.Args[len(os.Args)-1]
	command := os.Args[1]

	url, err := url.Parse(link)
	if err != nil {
		panic(err)
	}

	object, err := kinds.Fetch(url)
	if err != nil {
		panic(err)
	}

	if command == "raw" {
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(object); err != nil {
			panic(err)
		}
		return
	}

	str, _ := object.String()
	fmt.Println(str)
}
