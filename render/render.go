package render

import (
	"errors"
	"mimicry/hypertext"
	"mimicry/plaintext"
	"mimicry/gemtext"
	"mimicry/markdown"
	"strings"
)

// TODO: need to add a width parameter to all of this
func Render(text string, mediaType string) (string, error) {
	text = strings.Map(escapeControlCharacter, text)

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

func escapeControlCharacter(character rune) rune {
	if character >= 0 && character <= 31 && character != '\t' && character != '\n' && character != '\r' {
		return character + 0x2400
	}

	return character
}
