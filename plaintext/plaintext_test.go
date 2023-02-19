package plaintext

import (
	"testing"
	"mimicry/style"
	"mimicry/utils"
)

func TestBasic(t *testing.T) {
	input := `Yes, Jim, I found it under "http://www.w3.org/Addressing/",
but you can probably pick it up from <ftp://foo.example.com/rfc/>.
Note the warning in <http://www.ics.uci.edu/pub/ietf/uri/historical.html#WARNING>.`
	output, err := Render(input)
	if err != nil {
		panic(err)
	}

	expected := `Yes, Jim, I found it under "` +
		style.Link("http://www.w3.org/Addressing/") +
		`",
but you can probably pick it up from <` +
		style.Link("ftp://foo.example.com/rfc/") +
		`>.
Note the warning in <` +
		style.Link("http://www.ics.uci.edu/pub/ietf/uri/historical.html#WARNING") +
		`>.`

	utils.AssertEqual(expected, output, t)
}