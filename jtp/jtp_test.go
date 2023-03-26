package jtp

import (
	"testing"
	"mimicry/util"
	"net/url"
	"encoding/json"
	"os"
	"sync"
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
func TestBasic(t *testing.T) {
	accept := "application/activity+json"
	tolerated := []string{"application/json"}

	link, err := url.Parse("https://httpbin.org/redirect/20")
	if err != nil {
		panic(err)
	}

	var dict1, dict2 map[string]any
	var err1, err2 error
	var wg sync.WaitGroup; wg.Add(2); {
		go func() {
			dict1, err1 = Get(link, accept, tolerated, 20)
			wg.Done()
		}()
		go func() {
			dict2, err2 = Get(link, accept, tolerated, 20)
			wg.Done()
		}()
	}; wg.Wait()

	if err1 != nil {
		panic(err1)
	}

	if err2 != nil {
		panic(err2)
	}

	err = json.NewEncoder(os.Stdout).Encode(dict1)
	if err != nil {
		panic(err)
	}

	err = json.NewEncoder(os.Stdout).Encode(dict2)
	if err != nil {
		panic(err)
	}
}