package kinds

import (
	"errors"
	"net/url"
)

// TODO, add a verbose debugging output mode
// to debug problems that arise with this thing
// looping too much and whatnot

// `unstructured` is the JSON to construct from,
// source is where the JSON was received from,
// used to ensure the reponse is trustworthy
func Construct(unstructured Dict, source *url.URL) (Content, error) {
	kind, err := Get[string](unstructured, "type")
	if err != nil {
		return nil, err
	}

	// this requirement should be removed, and the below check
	// should be checking if only type or only type and id
	// are present on the element
	hasIdentifier := true
	id, err := GetURL(unstructured, "id")
	if err != nil {
		hasIdentifier = false
	}

	// if the JSON came from a source (e.g. inline in another collection), with a
	// different hostname than its ID, refetch
	// if the JSON only has two keys (type and id), refetch
	if (source != nil && id != nil) {
		if (source.Hostname() != id.Hostname()) || (len(unstructured) <= 2 && hasIdentifier) {
			return Fetch(id)
		}
	}

	switch kind {
	case "Article", "Audio", "Document", "Image", "Note", "Page", "Video":
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

	case "Application", "Group", "Organization", "Person", "Service":
		// TODO: nicer way to do this?
		actor := Actor{}
		actor = unstructured
		return actor, nil

	case "Link":
		link := Link{}
		link = unstructured
		return link, nil

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
