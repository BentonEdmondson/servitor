package pub

import (
	"strings"
	"mimicry/client"
)

func FetchUserInput(text string) Any {
	if strings.HasPrefix(text, "@") {
		link, err := client.ResolveWebfinger(text)
		if err != nil {
			return NewFailure(err)
		}
		return NewTangible(link, nil)
	}

	if strings.HasPrefix(text, "/") ||
		strings.HasPrefix(text, "./") ||
		strings.HasPrefix(text, "../") {
		object, err := client.FetchFromFile(text)
		if err != nil {
			return NewFailure(err)
		}
		return NewTangible(object, nil)
	}

	return NewTangible(text, nil)
}
