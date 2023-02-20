package render

import (
	"errors"
	"mimicry/hypertext"
	"mimicry/plaintext"
	"mimicry/gemtext"
	"mimicry/markdown"
	"strings"
	"unicode"
)

// TODO: need to add a width parameter to all of this
func Render(text string, mediaType string) (string, error) {
	text = strings.Map(dropControlCharacters, text)

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

func dropControlCharacters(character rune) rune {
	if unicode.IsControl(character) && character != '\t' && character != '\n' {
		return -1 // drop the character
	}

	return character
}
