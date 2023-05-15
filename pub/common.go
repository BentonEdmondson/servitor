package pub

import (
	"mimicry/object"
	"fmt"
	"errors"
	"net/url"
)

var (
	ErrWrongType = errors.New("item is the wrong type")
)

const (
	timeFormat = "3:04 pm on 2 Jan 2006"
)

/*
	This implements functions common to the different types.
	- getActors
	- getCollection
	- getActor
	- getPostOrActor
	- NewTangible

	// these will return an error on any problem
	- getBestLink, link impl will need the link, Rating(), mediatype, and be willing to take in Posts or Links
	- getFirstLinkShorthand
	- getBestLinkShorthand

	// used exclusively for attachments, honestly I
	// think it should probably return markup.
	// probably should actually be a function within 
	// Post
	- getLinks
*/

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

	// TODO: parallelize will probably require making fixed size
	// full width, swapping publics for nils, then later filtering
	// out the nils to reach a dynamic width
	output := []TangibleWithName{}
	for _, element := range list {
		if narrowed, ok := element.(string); ok {
			if narrowed == "https://www.w3.org/ns/activitystreams#Public" ||
			narrowed == "as:Public" ||
			narrowed == "Public" {
			continue
		}
		}

		fetched, err := NewActor(element, source)
		if err != nil {
			output = append(output, NewFailure(err))
		} else {
			output = append(output, fetched)
		}
	}
	return output
}

func getPostOrActor(o object.Object, key string, source *url.URL) Tangible {
	reference, err := o.GetAny(key)
	if err != nil {
		return NewFailure(err)
	}

	// TODO: add special case for lemmy where a json object with
	// type Create is automatically unwrapped right here

	var fetched Tangible
	fetched, err = NewActor(reference, source)
	if errors.Is(err, ErrWrongType) {
		fetched, err = NewPost(reference, source)
	}
	if err != nil {
		return NewFailure(err)
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