package markdown

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"servitor/hypertext"
)

var renderer = goldmark.New(goldmark.WithExtensions(extension.GFM))

func NewMarkup(text string) (*hypertext.Markup, []string, error) {
	var buf bytes.Buffer
	if err := renderer.Convert([]byte(text), &buf); err != nil {
		return nil, []string{}, err
	}
	output := buf.String()
	return hypertext.NewMarkup(output)
}
