package markdown

import (
    "bytes"
    "github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"mimicry/hypertext"
	"strings"
)

var renderer = goldmark.New(goldmark.WithExtensions(extension.GFM))

func Render(text string) (string, error) {
	var buf bytes.Buffer
	if err := renderer.Convert([]byte(text), &buf); err != nil {
		return "", nil
	}
	output := buf.String()
	rendered, err := hypertext.Render(output)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(rendered), nil
}