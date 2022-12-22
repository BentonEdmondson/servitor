package kinds

import (
	"errors"
	"net/url"
	"time"
)

// TODO throughout this file: attach the problematic object to the error

type Dict = map[string]any

func Get[T any](o Dict, key string) (T, error) {
	var zero T
	if value, ok := o[key]; !ok {
		return zero, errors.New("Object does not contain key " + key)
	} else if value, ok := value.(T); !ok {
		return zero, errors.New("Key " + key + " is not of the desired type")
	} else {
		return value, nil
	}
}

// some fields have "natural language values" meaning that I should check
// `contentMap[language]`, followed by `content`, followed by `contentMap["und"]`
// to find, e.g., the content of the post
// https://www.w3.org/TR/activitystreams-core/#naturalLanguageValues
func GetNatural(o Dict, key string, language string) (string, error) {
	values, valuesErr := Get[Dict](o, key+"Map")

	if valuesErr == nil {
		if value, err := Get[string](values, language); err == nil {
			return value, nil
		}
	}

	if value, err := Get[string](o, key); err == nil {
		return value, nil
	}

	if valuesErr == nil {
		if value, err := Get[string](values, "und"); err == nil {
			return value, nil
		}
	}

	return "", errors.New("Natural language key " + key + " is not correctly present in object")
}

// there may be a nice way to extract this logic out but for now it doesn't matter
func GetTime(o Dict, key string) (time.Time, error) {
	if value, err := Get[string](o, key); err != nil {
		return time.Time{}, err
	} else {
		return time.Parse(time.RFC3339, value)
	}
}
func GetURL(o Dict, key string) (*url.URL, error) {
	if value, err := Get[string](o, key); err != nil {
		return nil, err
	} else {
		return url.Parse(value)
	}
}

func GetContent[T Content](d Dict, key string) ([]T, error) {

	value, ok := d["attributedTo"]
	if !ok {
		return []T{}, nil
	}

	list := []any{}
	
	if valueList, isList := value.([]any); isList {
		list = valueList
	} else {
		list = []any{value}
	}
	
	output := []T{}
	
	for _, el := range list {
		switch narrowed := el.(type) {
		case Dict:
			// TODO: if source is absent, must refetch
			source, err := GetURL(d, "id")
			if err != nil { continue }
			resolved, err := Construct(narrowed, source)
			if err != nil { continue }
			asT, isT := resolved.(T)
			if !isT { continue }
			output = append(output, asT)
		case string:
			url, err := url.Parse(narrowed)
			if err != nil { continue }
			object, err := Fetch(url)
			if err != nil { continue }
			asT, isT := object.(T)
			if !isT { continue }
			output = append(output, asT)
		default: continue
		}
	}

	return output, nil
	
}
