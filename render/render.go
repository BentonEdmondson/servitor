package render

import (
	"strings"
	"errors"
	"fmt"
)

// Just use body as content because that only permits flow content
// https://stackoverflow.com/questions/15081119/any-way-to-use-html-parse-without-it-adding-nodes-to-make-a-well-formed-tree

func Render(text string, mediaType string) (string, error) {
	fmt.Println("started render")
	switch {
	case strings.Contains(mediaType, "text/plain"): 
		return text, nil
	case strings.Contains(mediaType, "text/html"):
		return renderHTML(text)
	default:
		return "", errors.New("Cannot render text of mime type " + mediaType)
	}
}
