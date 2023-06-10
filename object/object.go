package object

import (
	"errors"
	"fmt"
	"mimicry/mime"
	"net/url"
	"time"
)

type Object map[string]any

var (
	ErrKeyNotPresent = errors.New("key is not present")
	ErrKeyWrongType  = errors.New("key is incorrect type")
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
	return getPrimitive[string](o, key)
}

// TODO: should probably error for non-uints
func (o Object) GetNumber(key string) (uint64, error) {
	if number, err := getPrimitive[float64](o, key); err != nil {
		return 0, err
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

/* https://www.w3.org/TR/activitystreams-core/#naturalLanguageValues */
func (o Object) GetNatural(key string, language string) (string, error) {
	values, err := o.GetObject(key + "Map")
	hasMap := true
	if errors.Is(err, ErrKeyNotPresent) {
		hasMap = false
	} else if err != nil {
		return "", err
	}

	if hasMap {
		if value, err := values.GetString(language); err == nil {
			return value, nil
		} else if !errors.Is(err, ErrKeyNotPresent) {
			return "", fmt.Errorf("failed to extract from \"%s\": %w", key+"Map", err)
		}
	}

	if value, err := o.GetString(key); err == nil {
		return value, nil
	} else if !errors.Is(err, ErrKeyNotPresent) {
		return "", err
	}

	if hasMap {
		if value, err := values.GetString("und"); err == nil {
			return value, nil
		} else if !errors.Is(err, ErrKeyNotPresent) {
			return "", fmt.Errorf("failed to extract from \"%s\": %w", key+"Map", err)
		}
	}

	return "", fmt.Errorf("failed to extract natural \"%s\": %w", key, ErrKeyNotPresent)
}
