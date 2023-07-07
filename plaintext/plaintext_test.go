package plaintext

import (
	"mimicry/ansi"
	"mimicry/style"
	"testing"
)

func TestBasic(t *testing.T) {
	input := `Yes, Jim, I found it under "http://www.w3.org/Addressing/",
but you can probably pick it up from the store.
Note the warning in <http://www.ics.uci.edu/pub/ietf/uri/historical.html#WARNING>.`
	markup, links, err := NewMarkup(input)
	if err != nil {
		t.Fatal(err)
	}
	output := markup.Render(50)

	first := links[0]
	if first != "http://www.w3.org/Addressing/" {
		t.Fatalf("first uri should be http://www.w3.org/Addressing/ not %s", first)
	}

	second := links[1]
	if second != "http://www.ics.uci.edu/pub/ietf/uri/historical.html#WARNING" {
		t.Fatalf("first uri should be http://www.ics.uci.edu/pub/ietf/uri/historical.html#WARNING not %s", second)
	}

	expected := ansi.Wrap("Yes, Jim, I found it under \""+style.Link("http://www.w3.org/Addressing/", 1)+
		"\",\nbut you can probably pick it up from the store.\n"+
		"Note the warning in <"+style.Link("http://www.ics.uci.edu/pub/ietf/uri/historical.html#WARNING", 2)+">.", 50)

	if expected != output {
		t.Fatalf("expected markup to be %s not %s", expected, output)
	}
}
