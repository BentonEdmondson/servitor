package plaintext

import (
	"testing"
	"mimicry/style"
	"mimicry/util"
	"mimicry/ansi"
)

func TestBasic(t *testing.T) {
	input := `Yes, Jim, I found it under "http://www.w3.org/Addressing/",
but you can probably pick it up from <ftp://foo.example.com/rfc/>.
Note the warning in <http://www.ics.uci.edu/pub/ietf/uri/historical.html#WARNING>.`
	output, err := Render(input, 50)
	if err != nil {
		panic(err)
	}

	expected := ansi.Wrap("Yes, Jim, I found it under \"" + style.Link("http://www.w3.org/Addressing/") +
	"\",\nbut you can probably pick it up from <" + style.Link("ftp://foo.example.com/rfc/") +
	">.\nNote the warning in <" + style.Link("http://www.ics.uci.edu/pub/ietf/uri/historical.html#WARNING") + ">.", 50)

	util.AssertEqual(expected, output, t)
}