package markdown

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"mimicry/hypertext"
)

var renderer = goldmark.New(goldmark.WithExtensions(extension.GFM))

type Markup hypertext.Markup

func NewMarkup(text string) (*Markup, []string, error) {
	var buf bytes.Buffer
	if err := renderer.Convert([]byte(text), &buf); err != nil {
		return nil, []string{}, err
	}
	output := buf.String()
	hypertextMarkup, links, err := hypertext.NewMarkup(output)
	return (*Markup)(hypertextMarkup), links, err
}

func (m *Markup) Render(width int) string {
	return (*hypertext.Markup)(m).Render(width)
}
