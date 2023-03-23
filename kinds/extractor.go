package kinds

import (
	"errors"
	"net/url"
	"time"
)

// TODO: rename this file to dictionary.go
// TODO: rename Dict to Dictionary

// TODO: add a HasContent method. This is used when checking if
// content exists, so Actors can apply mediaType to summary,
// and Posts can apply it to url. In other situations (items vs
// orderedItems, name vs preferredName) attempting to fallback
// is always better than just failing, and for (absence of next/prev)
// if I don't have next, I have no other option, so it is 
// effectively the end of the line

// TODO: read through the spec's discussion on security

type Dict = map[string]any

func Get[T any](o Dict, key string) (T, error) {
	var zero T
	if value, ok := o[key]; !ok {
		return zero, errors.New("Object does not contain key " + key)
	} else if narrowed, ok := value.(T); !ok {
		return zero, errors.New("Key " + key + " is not of the desired type")
	} else {
		return narrowed, nil
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

// TODO: this needs to be switched over to using `GetItem`
/*
	`GetContent`
	For a given key, return all values of type T.
	Strings are interpreted as URLs and fetched.
	The Public address representations mentioned
	in §5.6 are ignored.
	
	Used for `Post.attributedTo` and `Post.inReplyTo`,
	for instance.
*/
func GetContent[T Content](d Dict, key string) ([]T, error) {
	values, err := GetList(d, key)
	if err != nil {
		return nil, err
	}
	
	output := []T{}
	
	for _, el := range values {
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
			// §5.6
			if narrowed == "https://www.w3.org/ns/activitystreams#Public" ||
				narrowed == "as:Public" || narrowed == "Public" { continue }
			url, err := url.Parse(narrowed)
			if err != nil { continue }
			object, err := FetchURL(url)
			if err != nil { continue }
			asT, isT := object.(T)
			if !isT { continue }
			output = append(output, asT)
		default: continue
		}
	}

	return output, nil
}

func GetItem[T Content](d Dict, key string) (T, error) {
	value, err := Get[any](d, key)
	if err != nil {
		var dummy T
		return dummy, err
	}

	source, _ := GetURL(d, "id")

	fetched, err := FetchUnknown(value, source)

	asT, isT := fetched.(T)

	if !isT {
		errors.New("Fetched " + key + " on " + source.String() + " is not of the desired type")
	}

	return asT, nil
}

/*
	`GetList`
	For a given key, return the value if it is a
	slice, if not, put it in a slice and return that.
*/
func GetList(d Dict, key string) ([]any, error) {
	value, err := Get[any](d, key)
	if err != nil { return []any{}, err }
	if asList, isList := value.([]any); isList {
		return asList, nil
	} else {
		return []any{value}, nil
	}
}

/*
	`GetLinksStrict`
	Returns a list
	of Links. Strings are interpreted as Links and
	are not fetched. If d.content is absent, d.mediaType
	is interpreted as applying to these strings.
	Non-string, non-Link elements are ignored.

	Used for `Post.url`.
*/
// TODO: for simplicity, make this a method of Post,
// it is easier to conceptualize when it works only on
// Posts, plus I can use my other post methods
func GetLinksStrict(d Dict, key string) ([]Link, error) {
	values, err := GetList(d, key)
	if err != nil {
		return nil, err
	}
	
	output := []Link{}

	// if content is absent and mediaType is present,
	// mediaType applies to the Links
	// name applies to the Links
	// nil/null represents absence
	var defaultMediaType any // (string | nil)
	mediaType, mediaTypeErr := Get[string](d, "mediaType")
	_, contentErr := GetNatural(d, "content", "en")
	if mediaTypeErr != nil || contentErr == nil {
		defaultMediaType = nil
	} else { defaultMediaType = mediaType }
	var defaultName any // (string | nil)
	if name, nameErr := GetNatural(d, "name", "en"); nameErr != nil {
		defaultName = nil
	} else { defaultName = name }

	for _, el := range values {
		switch narrowed := el.(type) {
		case string:
			output = append(output, Link{
				"type": "Link",
				"href": narrowed,
				"name": defaultName,
				"mediaType": defaultMediaType,
			})
		case Dict:
			// TODO: need to check this error?
			source, err := GetURL(d, "id")
			constructed, err := Construct(narrowed, source)
			if err != nil { continue }
			switch narrowedConstructed := constructed.(type) {
			case Link:
				output = append(output, narrowedConstructed)
			// TODO: ignore this case
			case Post:
				if postLink, err := narrowedConstructed.Link(); err != nil {
					output = append(output, postLink)
				} else { continue }
			default: continue
			}
		default: continue
		}
	}

	return output, nil
}

/*
	`GetLinksLenient`
	Similar to `GetLinksStrict`, but converts Posts
	to Links instead of ignoring them, and treats
	strings as URLs (not Links) and fetches them.

	Used for `Post.attachment`, `Actor.icon`, etc.
*/
func GetLinksLenient(d Dict, key string) ([]Link, error) {
	values, err := GetContent[Content](d, key)
	if err != nil {
		return []Link{}, err
	}

	output := []Link{}

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
