package gemtext

import (
	"servitor/style"
	"testing"
)

func TestBasic(t *testing.T) {
	input := `> blockquote

* bullet point

# large header
## smaller header
### smallest header

=> https://www.wikipedia.org/ Wikipedia is great!

=>http://example.org/

` + "```\ncode block\nhere\n```"
	markup, links, err := NewMarkup(input)
	if err != nil {
		t.Fatal(err)
	}

	if links[0] != "https://www.wikipedia.org/" {
		t.Fatalf("first link should be https://www.wikipedia.org/ not %s", links[0])
	}

	if links[1] != "http://example.org/" {
		t.Fatalf("second link should be http://example.org/ not %s", links[1])
	}

	output := markup.Render(50)
	expected := style.QuoteBlock("blockquote") + "\n\n" +
		style.Bullet("bullet point") + "\n\n" +
		style.Header("large header", 1) + "\n" +
		style.Header("smaller header", 2) + "\n" +
		style.Header("smallest header", 3) + "\n\n" +
		style.LinkBlock("Wikipedia is great!", 1) + "\n\n" +
		style.LinkBlock("http://example.org/", 2) + "\n\n" +
		style.CodeBlock("code block\nhere")

	if expected != output {
		t.Fatalf("expected %s not %s", expected, output)
	}
}
