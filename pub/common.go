package pub

import (
	"mimicry/object"
	"fmt"
	"errors"
	"net/url"
	"mimicry/client"
	"sync"
)

var (
	ErrWrongType = errors.New("item is the wrong type")
)

const (
	timeFormat = "3:04 pm on 2 Jan 2006"
)

type TangibleWithName interface {
	Tangible
	Name() string
}
func getActors(o object.Object, key string, source *url.URL) []TangibleWithName {
	list, err := o.GetList(key)
	if errors.Is(err, object.ErrKeyNotPresent) {
		return []TangibleWithName{}
	} else if err != nil {
		return []TangibleWithName{NewFailure(err)}
	}

	output := make([]TangibleWithName, len(list))
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
		if err != nil { return NewFailure(err) }
		if kind == "Create" {
			reference, err = o.GetAny("object")
			if err != nil { return NewFailure(err) }
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

func getCollection(o object.Object, key string, source *url.URL) (*Collection, error) {
	reference, err := o.GetAny(key)
	if err != nil {
		return nil, err
	}

	fetched, err := NewCollection(reference, source)
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

func NewTangible(input any, source *url.URL) Tangible {
	var fetched Tangible
	fetched, err := NewPost(input, source)

	if errors.Is(err, ErrWrongType) {
		fetched, err = NewActor(input, source)
	}

	if errors.Is(err, ErrWrongType) {
		fetched, err = NewActivity(input, source)
	}

	if errors.Is(err, ErrWrongType) {
		return NewFailure(err)
	}

	if err != nil {
		return NewFailure(err)
	}

	return fetched
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
		case object.Object:
			link, err := NewLink(narrowed)
			if err != nil {
				return nil, err
			}
			output[i] = link
		case string:
			link, err := NewLink(object.Object {
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
		return nil, err
	}
	links := make([]*Link, len(list))
	for i, element := range list {
		link, err := NewLink(element)
		if err != nil {
			return nil, err
		}
		links[i] = link
	}
	return links, nil
}

func getBestLink(o object.Object, key string, supertype string) (*Link, error) {
	links, err := getLinks(o, key)
	if err != nil { return nil, err }
	return SelectBestLink(links, supertype)
}
