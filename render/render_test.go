package render

import (
	"testing"
	"mimicry/style"
	"mimicry/util"
)

func TestControlCharacterEscapes(t *testing.T) {
	input := "Yes, \u0000Jim, I found it\tunder \u001Bhttp://www.w3.org/Addressing/"
	output, err := Render(input, "text/plain")
	if err != nil {
		panic(err)
	}

	expected := "Yes, \u2400Jim, I found it\tunder \u241b" +
		style.Link("http://www.w3.org/Addressing/")

	util.AssertEqual(expected, output, t)
}