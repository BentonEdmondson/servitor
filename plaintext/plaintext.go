package plaintext

import (
	"regexp"
	"mimicry/style"
	"strings"
	"mimicry/util"
)

func Render(text string, width int) (string, error) {
	/*
		Oversimplistic URL regexp based on RFC 3986, Appendix A
		It matches:
			<scheme>://<hierarchy>
		Where
			<scheme> is ALPHA *( ALPHA / DIGIT / "+" / "-" / "." )
			<hierarchy> is any of the characters listed in Appendix A:
				A-Z a-z 0-9 - . ? # / @ : [ ] % _ ~ ! $ & ' ( ) * + , ; =
	*/

	url := regexp.MustCompile(`[A-Za-z][A-Za-z0-9+\-.]*://[A-Za-z0-9.?#/@:%_~!$&'()*+,;=\[\]\-]+`)
	rendered := url.ReplaceAllStringFunc(text, style.Link)
	wrapped := util.Wrap(rendered, width)
	return strings.TrimSpace(wrapped), nil
}