package util

import (
	"testing"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
)

func AssertEqual(expected string, output string, t *testing.T) {
	if expected != output {
		t.Fatalf("Expected `%s` not `%s`\n", expected, output)
	}
}

func Wrap(text string, width int) string {
	if width < 1 {
		width = 1
	}
	return wrap.String(wordwrap.String(text, width), width)
}