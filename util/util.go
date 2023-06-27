package util

import (
	"testing"
)

// TODO: delete this function
func AssertEqual(expected string, output string, t *testing.T) {
	if expected != output {
		t.Fatalf("Expected `%s` not `%s`\n", expected, output)
	}
}
