package jtp

import (
	"testing"
	"mimicry/util"
	"net/url"
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
func TestRedirect(t *testing.T) {
	accept := "application/activity+json"
	tolerated := []string{"application/json"}

	link, err := url.Parse("https://httpbin.org/redirect/5")
	if err != nil {
		t.Fatalf("invalid url literal: %s", err)
	}

	_, actualLink, err := Get(link, accept, tolerated, 5)

	if err != nil {
		t.Fatalf("failed to preform request: %s", err)
	}

	if link.String() == actualLink.String() {
		t.Fatalf("failed to return the underlying url redirected to by %s", link.String())
	}
}

func TestBasic(t *testing.T) {
	accept := "application/activity+json"
	tolerated := []string{"application/json"}

	link, err := url.Parse("https://httpbin.org/get")
	if err != nil {
		t.Fatalf("invalid url literal: %s", err)
	}

	_, actualLink, err := Get(link, accept, tolerated, 20)

	if err != nil {
		t.Fatalf("failed to preform request: %s", err)
	}

	if link.String() != actualLink.String() {
		t.Fatalf("underlying url %s should match request url %s", actualLink.String(), link.String())
	}
}
