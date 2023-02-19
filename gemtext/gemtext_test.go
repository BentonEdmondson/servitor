package gemtext

import (
	"testing"
	"mimicry/style"
)

func assertEqual(expected string, output string, t *testing.T) {
	if expected != output {
		t.Fatalf("Expected `%s` not `%s`\n", expected, output)
	}
}

func TestBasic(t *testing.T) {
	input := `> blockquote

* bullet point

# large header
## smaller header
### smallest header

=> https://www.wikipedia.org/ Wikipedia is great!

=>http://example.org/

` + "```\ncode block\nhere\n```"
	output, err := Render(input)
	if err != nil {
		panic(err)
	}

	expected := style.QuoteBlock("blockquote") + "\n\n" +
		style.Bullet("bullet point") + "\n\n" +
		style.Header("large header", 1) + "\n" +
		style.Header("smaller header", 2) + "\n" +
		style.Header("smallest header", 3) + "\n\n" +
		style.LinkBlock("Wikipedia is great!") + "\n\n" +
		style.LinkBlock("http://example.org/") + "\n\n" +
		style.CodeBlock("code block\nhere")

	assertEqual(expected, output, t)
}