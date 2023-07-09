package pub

import (
	"errors"
	"fmt"
	"servitor/client"
	"servitor/object"
	"net/url"
	"sync"
	"time"
)

var (
	ErrWrongType = errors.New("item is the wrong type")
)

func getActors(o object.Object, key string, source *url.URL) []Tangible {
	list, err := o.GetList(key)
	if errors.Is(err, object.ErrKeyNotPresent) {
		return []Tangible{}
	} else if err != nil {
		return []Tangible{NewFailure(err)}
	}

	output := make([]Tangible, len(list))
	var wg sync.WaitGroup
	for i := range list {
		wg.Add(1)
		i := i
		go func() {
			fetched, err := NewActor(list[i], source)
			if err != nil {
				output[i] = NewFailure(err)
			} else {
				output[i] = fetched
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return output
}

func getPostOrActor(o object.Object, key string, source *url.URL) Tangible {
	reference, err := o.GetAny(key)
	if err != nil {
		return NewFailure(err)
	}

	/* For Lemmy compatibility, automatically unwrap if the entry is an
	   inline Create type */
	if asMap, ok := reference.(map[string]any); ok {
		o := object.Object(asMap)
		kind, err := o.GetString("type")
		if err != nil {
			return NewFailure(err)
		}
		if kind == "Create" {
			reference, err = o.GetAny("object")
			if err != nil {
				return NewFailure(err)
			}
		}
	}

	o, id, err := client.FetchUnknown(reference, source)
	if err != nil {
		return NewFailure(err)
	}

	var fetched Tangible
	var postErr, actorErr error
	fetched, postErr = NewPostFromObject(o, id)
	if errors.Is(postErr, ErrWrongType) {
		fetched, actorErr = NewActorFromObject(o, id)
		if errors.Is(actorErr, ErrWrongType) {
			return NewFailure(fmt.Errorf("%w, %w", postErr, actorErr))
		} else if actorErr != nil {
			return NewFailure(actorErr)
		}
	} else if postErr != nil {
		return NewFailure(postErr)
	}

	return fetched
}

func getCollection(o object.Object, key string, source *url.URL, construct func(any, *url.URL) Tangible) (*Collection, error) {
	reference, err := o.GetAny(key)
	if err != nil {
		return nil, err
	}

	fetched, err := NewCollection(reference, source, construct)
	if err != nil {
		return nil, err
	}
	return fetched, nil
}

func getActor(o object.Object, key string, source *url.URL) (*Actor, error) {
	reference, err := o.GetAny(key)
	if err != nil {
		return nil, err
	}

	fetched, err := NewActor(reference, source)
	if err != nil {
		return nil, err
	}
	return fetched, nil
}

func getAndFetchUnkown(o object.Object, key string, source *url.URL) (object.Object, *url.URL, error) {
	reference, err := o.GetAny(key)
	if err != nil {
		return nil, nil, err
	}

	return client.FetchUnknown(reference, source)
}

func NewTangible(input any, source *url.URL) Tangible {
	fetched := New(input, source)
	if tangible, ok := fetched.(Tangible); ok {
		return tangible
	}
	return NewFailure(errors.New("item is a collection"))
}

func New(input any, source *url.URL) any {
	o, id, err := client.FetchUnknown(input, source)
	if err != nil {
		return NewFailure(err)
	}

	var result any

	result, err = NewActorFromObject(o, id)
	if err == nil {
		return result
	} else if !errors.Is(err, ErrWrongType) {
		return NewFailure(err)
	}

	result, err = NewPostFromObject(o, id)
	if err == nil {
		return result
	} else if !errors.Is(err, ErrWrongType) {
		return NewFailure(err)
	}

	result, err = NewActivityFromObject(o, id)
	if err == nil {
		return result
	} else if !errors.Is(err, ErrWrongType) {
		return NewFailure(err)
	}

	result, err = NewCollectionFromObject(o, id, NewTangible)
	if err == nil {
		return result
	} else if !errors.Is(err, ErrWrongType) {
		return NewFailure(err)
	}

	return NewFailure(errors.New("item is of unrecognized type"))
}

/*
"Shorthand" just means individual strings are converted into Links
*/
func getLinksShorthand(o object.Object, key string) ([]*Link, error) {
	list, err := o.GetList(key)
	if err != nil {
		return nil, err
	}

	output := make([]*Link, len(list))

	for i, element := range list {
		switch narrowed := element.(type) {
		case map[string]any:
			link, err := NewLink(narrowed)
			if err != nil {
				return nil, err
			}
			output[i] = link
		case string:
			link, err := NewLink(object.Object{
				"type": "Link",
				"href": narrowed,
			})
			if err != nil {
				return nil, err
			}
			output[i] = link
		default:
			return nil, fmt.Errorf("can't convert a %T into a Link", element)
		}
	}
	return output, nil
}

func getBestLinkShorthand(o object.Object, key string, supertype string) (*Link, error) {
	links, err := getLinksShorthand(o, key)
	if err != nil {
		return nil, err
	}
	return SelectBestLink(links, supertype)
}

func getFirstLinkShorthand(o object.Object, key string) (*Link, error) {
	links, err := getLinksShorthand(o, key)
	if err != nil {
		return nil, err
	}
	return SelectFirstLink(links)
}

func getLinks(o object.Object, key string) ([]*Link, error) {
	list, err := o.GetList(key)
	if err != nil {
		return []*Link{}, err
	}
	links := make([]*Link, len(list))
	for i, element := range list {
		link, err := NewLink(element)
		if err != nil {
			return []*Link{}, err
		}
		links[i] = link
	}
	return links, nil
}

func getBestLink(o object.Object, key string, supertype string) (*Link, error) {
	links, err := getLinks(o, key)
	if err != nil {
		return nil, err
	}
	return SelectBestLink(links, supertype)
}

func ago(t time.Time) string {
	duration := time.Since(t)

	if days := int(duration.Hours() / 24); days > 1 {
		return fmt.Sprintf("%d days ago", int(days))
	} else if days == 1 {
		return "1 day ago"
	}

	if hours := int(duration.Hours()); hours > 1 {
		return fmt.Sprintf("%d hours ago", int(hours))
	} else if hours == 1 {
		return "1 hour ago"
	}

	if minutes := int(duration.Minutes()); minutes > 1 {
		return fmt.Sprintf("%d minutes ago", int(minutes))
	} else if minutes == 1 {
		return "1 minute ago"
	}

	return "seconds ago"
}
