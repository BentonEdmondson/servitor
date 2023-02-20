package plaintext

import (
	"testing"
	"mimicry/style"
	"mimicry/util"
)

func TestBasic(t *testing.T) {
	input := `Yes, Jim, I found it under "http://www.w3.org/Addressing/",
but you can probably pick it up from <ftp://foo.example.com/rfc/>.
Note the warning in <http://www.ics.uci.edu/pub/ietf/uri/historical.html#WARNING>.`
	output, err := Render(input, 50)
	if err != nil {
		panic(err)
	}

	expected := "Yes, Jim, I found it under\n\"" +
		style.Link("http://www.w3.org/Addressing/") +
		"\",\nbut you can probably pick it up from\n<" +
		style.Link("ftp://foo.example.com/rfc/") +
		">.\nNote the warning in\n<" +
		style.Link("http://www.ics.uci.edu/pub/ietf/uri/historical.ht\nml#WARNING") +
		">."

	util.AssertEqual(expected, output, t)
}