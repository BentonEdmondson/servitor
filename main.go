package main

import (
	"encoding/json"
	"fmt"
	"mimicry/kinds"
	"mimicry/request"
	"net/url"
	"os"
)

func main() {
	// I need to figure out the higher level abstractions
	// a package with a function that takes a url as string
	// and returns an Activity, Actor, Collection, or Post
	// really it will return a Thing interface, which implements
	// Kind, Identifier, String

	// the request function will need to disable bs like cookies,
	// etc, enable caching, set Accept header, check the header and
	// status code after receiving the request, parse the json with
	// strict validation, look at type to determine what to construct,
	// then return it

	// Other types I need to make are Link and Markup

	// TODO: maybe make a package called onboard that combines
	// request, extractor, and create
	// onboard.Fetch, onboard.Construct, onboard.Get, etc

	link := os.Args[len(os.Args)-1]
	command := os.Args[1]

	url, err := url.Parse(link)
	if err != nil {
		panic(err)
	}

	unstructured, err := request.Fetch(url)
	if err != nil {
		panic(err)
	}

	if command == "raw" {
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(unstructured); err != nil {
			panic(err)
		}
		return
	}

	object, err := kinds.Create(unstructured, url)
	if err != nil {
		panic(err)
	}

	fmt.Println(object.String())
}
