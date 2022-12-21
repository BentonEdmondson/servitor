package render

import (
	"strings"
	"errors"
)

func Render(text string, kind string) (string, error) {
	switch {
	case strings.Contains(kind, "text/plain"): 
		return text, nil
	case strings.Contains(kind, "text/html"):
		node, err := html.Parse(text)
		if err == nil {
			return "", err
		}
		return renderHTML(node), nil
	default:
		return "", errors.New("Cannot render text of mime type %s", kind)
}
