package object

import (
	"errors"
	"fmt"
	"math"
	"mimicry/ansi"
	"mimicry/gemtext"
	"mimicry/hypertext"
	"mimicry/markdown"
	"mimicry/mime"
	"mimicry/plaintext"
	"net/url"
	"time"
)

type Object map[string]any

var (
	ErrKeyNotPresent = errors.New("key is not present")
	ErrKeyWrongType  = errors.New("value is incorrect type")
)

/* Go doesn't allow generic methods */
func getPrimitive[T any](o Object, key string) (T, error) {
	var zero T
	if value, ok := o[key]; !ok || value == nil {
		return zero, fmt.Errorf("failed to extract \"%s\": %w", key, ErrKeyNotPresent)
	} else if narrowed, ok := value.(T); !ok {
		return zero, fmt.Errorf("failed to extract \"%s\": %w: is %T", key, ErrKeyWrongType, value)
	} else {
		return narrowed, nil
	}
}

func (o Object) GetAny(key string) (any, error) {
	return getPrimitive[any](o, key)
}

func (o Object) GetString(key string) (string, error) {
	value, err := getPrimitive[string](o, key)
	if err != nil {
		return "", err
	}
	value = ansi.Scrub(value)
	if value == "" {
		return "", ErrKeyNotPresent
	}
	return value, nil
}

func (o Object) GetNumber(key string) (uint64, error) {
	if number, err := getPrimitive[float64](o, key); err != nil {
		return 0, err
	} else if number != math.Trunc(number) {
		return 0, fmt.Errorf("failed to extract \"%s\": value is not a non-integer number", key)
	} else {
		return uint64(number), nil
	}
}

func (o Object) GetObject(key string) (Object, error) {
	return getPrimitive[map[string]any](o, key)
}

func (o Object) GetList(key string) ([]any, error) {
	if value, err := o.GetAny(key); err != nil {
		return nil, err
	} else if asList, isList := value.([]any); isList {
		return asList, nil
	} else {
		return []any{value}, nil
	}
}

func (o Object) GetTime(key string) (time.Time, error) {
	if value, err := o.GetString(key); err != nil {
		return time.Time{}, err
	} else {
		timestamp, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse time \"%s\": %w", key, err)
		}
		return timestamp, nil
	}
}

func (o Object) GetURL(key string) (*url.URL, error) {
	if value, err := o.GetString(key); err != nil {
		return nil, err
	} else {
		address, err := url.Parse(value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL \"%s\": %w", key, err)
		}
		return address, nil
	}
}

func (o Object) GetMediaType(key string) (*mime.MediaType, error) {
	if value, err := o.GetString(key); err != nil {
		return nil, err
	} else {
		mediaType, err := mime.Parse(value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse mime type \"%s\": %w", key, err)
		}
		return mediaType, nil
	}
}

type Markup interface {
	Render(width int) string
}

func (o Object) GetMarkup(contentKey string, mediaTypeKey string) (Markup, []string, error) {
	content, err := o.GetString(contentKey)
	if err != nil {
		return nil, nil, err
	}
	mediaType, err := o.GetMediaType(mediaTypeKey)
	if errors.Is(err, ErrKeyNotPresent) {
		mediaType = mime.Default()
	} else if err != nil {
		return nil, nil, err
	}

	switch mediaType.Essence {
	case "text/plain":
		return plaintext.NewMarkup(content)
	case "text/html":
		return hypertext.NewMarkup(content)
	case "text/gemini":
		return gemtext.NewMarkup(content)
	case "text/markdown":
		return markdown.NewMarkup(content)
	default:
		return nil, nil, errors.New("cannot render text of mime type " + mediaType.Essence)
	}
}
