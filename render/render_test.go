package render

import (
	"testing"
	"mimicry/style"
	"mimicry/util"
)

func TestControlCharacterEscapes(t *testing.T) {
	input := "Yes, \u0000Jim, I\nfound it\tunder \u001Bhttp://www.w3.org/Addressing/"
	output, err := Render(input, "text/plain", 50)
	if err != nil {
		panic(err)
	}

	expected := "Yes, Jim, I\nfound it\tunder " +
		style.Link("http://www.w3.org/Addressing/")

	util.AssertEqual(expected, output, t)
}