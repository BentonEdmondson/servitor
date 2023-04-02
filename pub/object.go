package pub

import (
	"errors"
	"net/url"
	"time"
	"mimicry/jtp"
	"mimicry/render"
	"fmt"
)

type Object map[string]any

/*
	These are helper functions that really should be methods, but
	Go does not allow generic methods.
*/
func getPrimitive[T any](o Object, key string) (T, error) {
	var zero T
	if value, ok := o[key]; !ok {
		return zero, fmt.Errorf("object does not contain key \"" + key + "\": %v", o)
	} else if narrowed, ok := value.(T); !ok {
		return zero, errors.New("key " + key + " is not of the desired type")
	} else {
		return narrowed, nil
	}
}
func getItem[T Item](o Object, key string) (T, error) {
	value, err := getPrimitive[any](o, key)
	if err != nil {
		return *new(T), err
	}
	source, _ := o.GetURL("id")
	fetched, err := FetchUnknown(value, source)
	if err != nil { return *new(T), err }
	asT, isT := fetched.(T)
	if !isT {
		errors.New("fetched " + key + " is not of the desired type")
	}
	return asT, nil
}
func getItems[T Item](o Object, key string) ([]T, error) {
	values, err := o.GetList(key)
	if err != nil {
		return nil, err
	}
	source, _ := o.GetURL("id")
	output := make([]T, 0, len(values))
	for _, el := range values {
		resolved, err := FetchUnknown(el, source)
		if err != nil { continue }
		asT, isT := resolved.(T)
		if !isT { continue }
		output = append(output, asT)
	}
	return output, nil
}

/* Various methods for getting basic information from the Object */
func (o Object) GetString(key string) (string, error) {
	return getPrimitive[string](o, key)
}
func (o Object) GetNumber(key string) (uint64, error) {
	if number, err := getPrimitive[float64](o, key); err != nil {
		return 0, err
	} else {
		return uint64(number), nil
	}
}
func (o Object) GetObject(key string) (Object, error) {
	return getPrimitive[Object](o, key)
}
func (o Object) GetList(key string) ([]any, error) {
	if value, err := getPrimitive[any](o, key); err != nil {
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
		return time.Parse(time.RFC3339, value)
	}
}
func (o Object) GetURL(key string) (*url.URL, error) {
	if value, err := o.GetString(key); err != nil {
		return nil, err
	} else {
		return url.Parse(value)
	}
}
func (o Object) GetMediaType(key string) (*jtp.MediaType, error) {
	if value, err := o.GetString(key); err != nil {
		return nil, err
	} else {
		return jtp.ParseMediaType(value)
	}
}
/* https://www.w3.org/TR/activitystreams-core/#naturalLanguageValues */
func (o Object) GetNatural(key string, language string) (string, error) {
	values, valuesErr := o.GetObject(key+"Map")
	if valuesErr == nil {
		if value, err := values.GetString(language); err == nil {
			return value, nil
		}
	}
	if value, err := o.GetString(key); err == nil {
		return value, nil
	}
	if valuesErr == nil {
		if value, err := values.GetString("und"); err == nil {
			return value, nil
		}
	}
	return "", errors.New("natural language key " + key + " is not correctly present in object")
}

/* Methods for getting various Items from the Object */
func (o Object) GetActors(key string) ([]Actor, error) {
	return getItems[Actor](o, key)
}
func (o Object) GetPost(key string) (Post, error) {
	return getItem[Post](o, key)
}
// func (o Object) GetActivity(key string) (Activity, error) {
// 	return getItem[Activity](o, key)
// }
func (o Object) GetCollection(key string) (Collection, error) {
	return getItem[Collection](o, key)
}
func (o Object) GetItems(key string) ([]Item, error) {
	return getItems[Item](o, key)
}

/*
	Fetches strings as URLs, converts Posts to Links, and
		ignores non-Link non-Post non-string elements.
	Used for `Post.attachment`, `Actor.icon`, etc.
*/
func (o Object) GetLinks(key string) ([]Link, error) {
	values, err := o.GetItems(key)
	if err != nil {
		return []Link{}, err
	}
	output := make([]Link, 0, len(values))
	for _, el := range values {
		switch narrowed := el.(type) {
		case Link:
			output = append(output, narrowed)
		case Post:
			if link, err := narrowed.Link(); err == nil {
				output = append(output, link)
			} else { continue }
		default: continue
		}
	}
	return output, nil
}

func (o Object) Has(key string) bool {
	_, present := o[key]
	return present
}
func (o Object) HasNatural(key string) bool {
	return o.Has(key) || o.Has(key+"Map")
}

func (o Object) Render(contentKey string, langKey string, mediaTypeKey string, width int) (string, error) {
	body, err := o.GetNatural(contentKey, langKey)
	if err != nil {
		return "", err
	}
	mediaType := &jtp.MediaType{
		Supertype: "text",
		Subtype: "html",
		Full: "text/html",
	}
	if o.Has("mediaType") {
		mediaType, err = o.GetMediaType(mediaTypeKey)
		if err != nil { return "", err }
	}
	return render.Render(body, mediaType, width)
}
