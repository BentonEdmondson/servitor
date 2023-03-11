package jtp

import (
	"testing"
	// "net/url"
	// "encoding/json"
	// "os"
	"mimicry/util"
)

func TestStatusLineNoInfo(t *testing.T) {
	test := "HTTP/1.1 200\r\n"
	status, err := parseStatusLine(test)
	if err != nil {
		panic(err)
	}
	util.AssertEqual("200", status, t)
}

// TODO: put this behind an --online flag or figure out
// how to nicely do offline tests
// func TestBasic(t *testing.T) {
// 	link, err := url.Parse("https://pixelfed.social/p/KLeon/535879788929882013")
// 	if err != nil {
// 		panic(err)
// 	}

// 	accept := `application/ld+json; profile="https://www.w3.org/ns/activitystreams", application/activity+json`
// 	tolerated := []string{
// 		"application/activity+json",
// 		"application/ld+json",
// 		"application/json",
// 	}

// 	dict, err := Get(link, accept, tolerated)
// 	if err != nil {
// 		panic(err)
// 	}

// 	err = json.NewEncoder(os.Stdout).Encode(dict)
// 	if err != nil {
// 		panic(err)
// 	}
// }