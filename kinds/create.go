package kinds

import (
	"errors"
	"net/url"
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

// `unstructured` is the JSON to construct from,
// source is where the JSON was received from,
// used to ensure the reponse is trustworthy
func Construct(unstructured JSON, source *url.URL) (Object, error) {
	kind, err := Get[string](unstructured, "type")
	if err != nil {
		return nil, err
	}

	id, err := GetURL(unstructured, "id")
	if err != nil {
		return nil, err
	}

	// if the JSON came from a source (e.g. inline in another collection), with a
	// different hostname than its ID, refetch
	// if the JSON only has two keys (type and id), refetch
	if source != nil && source.Hostname() != id.Hostname() || len(unstructured) <= 2 {
		return Fetch(id)
	}

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
