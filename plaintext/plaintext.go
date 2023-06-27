package plaintext

import (
	"mimicry/ansi"
	"mimicry/style"
	"regexp"
	"strings"
)

type Markup string

func NewMarkup(text string) (*Markup, []string, error) {
	rendered, links := renderWithLinks(text, 80)
	return (*Markup)(&rendered), links, nil
}

func (m Markup) Render(width int) string {
	rendered, _ := renderWithLinks(string(m), width)
	return rendered
}

func renderWithLinks(text string, width int) (string, []string) {
	/*
		Oversimplistic URL regexp based on RFC 3986, Appendix A
		It matches:
			<scheme>://<hierarchy>
		Where
			<scheme> is ALPHA *( ALPHA / DIGIT / "+" / "-" / "." )
			<hierarchy> is any of the characters listed in Appendix A:
				A-Z a-z 0-9 - . ? # / @ : [ ] % _ ~ ! $ & ' ( ) * + , ; =
	*/

	links := []string{}

	url := regexp.MustCompile(`[A-Za-z][A-Za-z0-9+\-.]*://[A-Za-z0-9.?#/@:%_~!$&'()*+,;=\[\]\-]+`)
	rendered := url.ReplaceAllStringFunc(text, func(link string) string {
		links = append(links, link)

		// TODO: this will be superscripted
		return style.Link(link, len(links))
	})
	wrapped := ansi.Wrap(rendered, width)
	return strings.Trim(wrapped, "\n"), links
}
