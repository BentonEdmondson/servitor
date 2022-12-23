package kinds

import (
	"net/url"
	"strings"
)

type Link Dict

// one of these should be omitted so
// Link isn't Content
func (l Link) Kind() (string, error) {
	return "link", nil
}
func (l Link) Category() string {
	return "link"
}

func (l Link) MediaType() (string, error) {
	return Get[string](l, "mediaType")
}

func (l Link) URL() (*url.URL, error) {
	return GetURL(l, "href")
}

func (l Link) Alt() (string, error) {
	alt, err := Get[string](l, "name")
	return strings.TrimSpace(alt), err
}

func (l Link) Identifier() (*url.URL, error) {
	return nil, nil
}

// TODO: update of course to be nice markup of some sort
func (l Link) String() (string, error) {
	output := ""

	if alt, err := l.Alt(); err == nil {
		output += alt
	} else if url, err := l.URL(); err == nil {
		output += url.String()
	}

	if mediaType, err := l.MediaType(); err == nil {
		output += " (" + mediaType + ")"
	}

	return output, nil
}
