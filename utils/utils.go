package utils

import (
	"testing"
)

func AssertEqual(expected string, output string, t *testing.T) {
	if expected != output {
		t.Fatalf("Expected `%s` not `%s`\n", expected, output)
	}
}