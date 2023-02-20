package render

import (
	"errors"
	"mimicry/hypertext"
	"mimicry/plaintext"
	"mimicry/gemtext"
	"mimicry/markdown"
)

func Render(text string, mediaType string) (string, error) {
	switch {
	case mediaType == "text/plain": 
		return plaintext.Render(text)
	case mediaType == "text/html":
		return hypertext.Render(text)
	case mediaType == "text/gemini":
		return gemtext.Render(text)
	case mediaType == "text/markdown":
		return markdown.Render(text)
	default:
		return "", errors.New("Cannot render text of mime type " + mediaType)
	}
}
