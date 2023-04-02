package render

import (
	"errors"
	"mimicry/hypertext"
	"mimicry/plaintext"
	"mimicry/gemtext"
	"mimicry/markdown"
	"strings"
	"unicode"
	"mimicry/jtp"
)

// TODO: perhaps `dropControlCharacters` should happen to all
//	`getNatural` strings when they are pulled from the JSON
// TODO: need to add a width parameter to all of this
func Render(text string, mediaType *jtp.MediaType, width int) (string, error) {
	text = strings.Map(dropControlCharacters, text)

	switch {
	case mediaType.Full == "text/plain": 
		return plaintext.Render(text, width)
	case mediaType.Full == "text/html":
		return hypertext.Render(text, width)
	case mediaType.Full == "text/gemini":
		return gemtext.Render(text, width)
	case mediaType.Full == "text/markdown":
		return markdown.Render(text, width)
	default:
		return "", errors.New("cannot render text of mime type " + mediaType.Full)
	}
}

func dropControlCharacters(character rune) rune {
	if unicode.IsControl(character) && character != '\t' && character != '\n' {
		return -1 // drop the character
	}

	return character
}
