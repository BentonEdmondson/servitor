package pub

import (
	"servitor/client"
	"strings"
)

func FetchUserInput(text string) Any {
	if strings.HasPrefix(text, "@") || strings.HasPrefix(text, "!") {
		link, err := client.ResolveWebfinger(text[1:])
		if err != nil {
			return NewFailure(err)
		}
		return New(link, nil)
	}

	if strings.HasPrefix(text, "/") ||
		strings.HasPrefix(text, "./") ||
		strings.HasPrefix(text, "../") {
		object, err := client.FetchFromFile(text)
		if err != nil {
			return NewFailure(err)
		}
		return New(object, nil)
	}

	return New(text, nil)
}
