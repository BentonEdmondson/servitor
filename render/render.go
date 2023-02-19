package render

import (
	"strings"
	"errors"
	"fmt"
	"mimicry/hypertext"
	"mimicry/plaintext"
)

// TODO: need to actually parse mediaType, not use `Contains`
func Render(text string, mediaType string) (string, error) {
	fmt.Println("started render")
	switch {
	case strings.Contains(mediaType, "text/plain"): 
		return plaintext.Render(text)
	case strings.Contains(mediaType, "text/html"):
		return hypertext.Render(text)
	default:
		return "", errors.New("Cannot render text of mime type " + mediaType)
	}
}
