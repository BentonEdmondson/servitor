package kinds

// TODO: rename this to `construct`
// TODO: I think this should be moved to
// package request, then Fetch will return an Object
// directly by calling Create, and Create will
// still work fine

import (
	"errors"
	"net/url"
	"mimicry/shared"
	"mimicry/request"
)

type Object interface {
	String() string
	Kind() (string, error)
	Identifier() (*url.URL, error)
	Category() string
}

// TODO, add a verbose debugging output mode
// to debug problems that arise with this thing
// looping too much and whatnot

// source is where it came from, if source is different
// from the element's id, it will be refetched

// maybe change back to taking in a unstructured shared.JSON
func Create(input any, source *url.URL) (Object, error) {
	unstructured, ok := input.(shared.JSON)
	if !ok {
		return nil, errors.New("Cannot construct with a non-object JSON")
	}

	kind, err := shared.Get[string](unstructured, "type")
	if err != nil {
		return nil, err
	}

	id, err := shared.GetURL(unstructured, "id")
	if err != nil {
		return nil, err
	}

	// if the JSON came from a source (e.g. inline in another collection), with a
	// different hostname than its ID, refetch
	// if the JSON only has two keys (type and id), refetch
	if source != nil && source.Hostname() != id.Hostname() || len(unstructured) <= 2 {
		response, err := request.Fetch(id)
		if err != nil {
			return nil, err
		}
		return Create(response, nil)
	}

	// TODO: if the only keys are id and type,
	// you need to do a fetch to get the other keys

	switch kind {
	case "Article":
		fallthrough
	case "Audio":
		fallthrough
	case "Document":
		fallthrough
	case "Image":
		fallthrough
	case "Note":
		fallthrough
	case "Page":
		fallthrough
	case "Video":
		// TODO: figure out the way to do this directly
		post := Post{}
		post = unstructured
		return post, nil

	// case "Create":
	// 	fallthrough
	// case "Announce":
	// 	fallthrough
	// case "Dislike":
	// 	fallthrough
	// case "Like":
	// 	fallthrough
	// case "Question":
	// 	return Activity{unstructured}, nil

	case "Application":
		fallthrough
	case "Group":
		fallthrough
	case "Organization":
		fallthrough
	case "Person":
		fallthrough
	case "Service":
		// TODO: nicer way to do this?
		actor := Actor{}
		actor = unstructured
		return actor, nil

	// case "Link":
	// 	return Link{unstructured}, nil

	// case "Collection":
	// 	fallthrough
	// case "OrderedCollection":
	// 	return Collection{unstructured}, nil

	// case "CollectionPage":
	// 	fallthrough
	// case "OrderedCollectionPage":
	// 	return CollectionPage{unstructured}, nil

	default:
		return nil, errors.New("Object of Type " + kind + " unsupported")
	}
}
