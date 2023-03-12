package jtp

import (
	"testing"
	"mimicry/util"
	// "net/url"
	// "encoding/json"
	// "os"
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
// 	link, err := url.Parse("https://httpbin.org/redirect/20")
// 	if err != nil {
// 		panic(err)
// 	}

// 	dict, err := Get(link, 20)
// 	if err != nil {
// 		panic(err)
// 	}

// 	err = json.NewEncoder(os.Stdout).Encode(dict)
// 	if err != nil {
// 		panic(err)
// 	}
// }