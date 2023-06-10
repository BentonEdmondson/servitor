package mime

import (
	"errors"
	"regexp"
)

type MediaType struct {
	Essence   string
	Supertype string
	Subtype   string
}

/*
See: https://httpwg.org/specs/rfc9110.html#media.type
*/
var re = regexp.MustCompile(`(?s)^(([!#$%&'*+\-.^_\x60|~a-zA-Z0-9]+)/([!#$%&'*+\-.^_\x60|~a-zA-Z0-9]+)).*$`)

func Default() *MediaType {
	return &MediaType{
		Essence:   "text/html",
		Supertype: "text",
		Subtype:   "html",
	}
}

func Parse(input string) (*MediaType, error) {
	matches := re.FindStringSubmatch(input)

	if len(matches) != 4 {
		return nil, errors.New(`"` + input + `" is not a valid media type`)
	}

	return &MediaType{
		Essence:   matches[1],
		Supertype: matches[2],
		Subtype:   matches[3],
	}, nil
}

func (m *MediaType) Update(input string) error {
	parsed, err := Parse(input)
	if err != nil {
		return err
	}
	*m = *parsed
	return nil
}

func (m *MediaType) Matches(mediaTypes []string) bool {
	for _, mediaType := range mediaTypes {
		if m.Essence == mediaType {
			return true
		}
	}
	return false
}
