package pub

import (
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"servitor/mime"
	"servitor/object"
	"net/url"
	"strings"
)

type Link struct {
	kind         string
	mediaType    *mime.MediaType
	mediaTypeErr error
	uri          *url.URL
	uriErr       error
	alt          string
	altErr       error
	height       uint64
	heightErr    error
	width        uint64
	widthErr     error
}

func NewLink(input any) (*Link, error) {
	l := &Link{}

	asMap, ok := input.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("can't turn non-object %T into Link", input)
	}
	o := object.Object(asMap)

	var err error
	if l.kind, err = o.GetString("type"); err != nil {
		return nil, err
	}

	if !slices.Contains([]string{
		"Link", "Audio", "Document", "Image", "Video",
	}, l.kind) {
		return nil, fmt.Errorf("%w: %s is not a Link", ErrWrongType, l.kind)
	}

	if l.kind == "Link" {
		l.uri, l.uriErr = o.GetURL("href")
		l.height, l.heightErr = o.GetNumber("height")
		l.width, l.widthErr = o.GetNumber("width")
	} else {
		l.uri, l.uriErr = o.GetURL("url")
		l.heightErr = object.ErrKeyNotPresent
		l.widthErr = object.ErrKeyNotPresent
	}

	l.mediaType, l.mediaTypeErr = o.GetMediaType("mediaType")
	l.alt, l.altErr = o.GetString("name")

	return l, nil
}

func (l *Link) Alt() (string, error) {
	if l.altErr == nil {
		return l.alt, nil
	} else if errors.Is(l.altErr, object.ErrKeyNotPresent) {
		if l.uriErr == nil {
			return l.uri.String(), nil
		} else {
			return "", l.uriErr
		}
	} else {
		return "", l.altErr
	}
}

func (l *Link) rating() (uint64, error) {
	var height, width uint64
	if l.heightErr == nil {
		height = l.height
	} else if errors.Is(l.heightErr, object.ErrKeyNotPresent) {
		height = 1
	} else {
		return 0, l.heightErr
	}
	if l.widthErr == nil {
		width = l.width
	} else if errors.Is(l.widthErr, object.ErrKeyNotPresent) {
		width = 1
	} else {
		return 0, l.widthErr
	}
	return height * width, nil
}

func SelectBestLink(links []*Link, supertype string) (*Link, error) {
	if len(links) == 0 {
		return nil, errors.New("can't select best link of type " + supertype + "/* from an empty list")
	}

	bestLink := links[0]

	// TODO: loop through once and validate errors, then proceed assuming no errors

	for _, thisLink := range links[1:] {
		var bestLinkSupertypeMatches bool
		if errors.Is(bestLink.mediaTypeErr, object.ErrKeyNotPresent) {
			bestLinkSupertypeMatches = false
		} else if bestLink.mediaTypeErr != nil {
			return nil, bestLink.mediaTypeErr
		} else {
			bestLinkSupertypeMatches = bestLink.mediaType.Supertype == supertype
		}

		var thisLinkSuperTypeMatches bool
		if errors.Is(thisLink.mediaTypeErr, object.ErrKeyNotPresent) {
			thisLinkSuperTypeMatches = false
		} else if thisLink.mediaTypeErr != nil {
			return nil, thisLink.mediaTypeErr
		} else {
			thisLinkSuperTypeMatches = thisLink.mediaType.Supertype == supertype
		}

		if thisLinkSuperTypeMatches && !bestLinkSupertypeMatches {
			bestLink = thisLink
			continue
		} else if !thisLinkSuperTypeMatches && bestLinkSupertypeMatches {
			continue
		} else {
			thisRating, err := thisLink.rating()
			if err != nil {
				return nil, err
			}
			bestRating, err := bestLink.rating()
			if err != nil {
				return nil, err
			}
			if thisRating > bestRating {
				bestLink = thisLink
				continue
			}
		}
	}

	return bestLink, nil
}

func (l *Link) Select() (string, *mime.MediaType, bool) {
	return l.SelectWithDefaultMediaType(mime.Unknown())
}

func (l *Link) SelectWithDefaultMediaType(defaultMediaType *mime.MediaType) (string, *mime.MediaType, bool) {
	if l.uriErr != nil {
		return "", nil, false
	}

	/* I suppress this error here because it is shown in the alt text */
	if l.mediaTypeErr == nil {
		return l.uri.String(), l.mediaType, true
	}

	if l.kind == "Audio" || l.kind == "Video" || l.kind == "Image" {
		return l.uri.String(), mime.UnknownSubtype(strings.ToLower(l.kind)), true
	}

	return l.uri.String(), defaultMediaType, true
}

func SelectFirstLink(links []*Link) (*Link, error) {
	if len(links) == 0 {
		return &Link{}, errors.New("can't select first Link from an empty list of links")
	} else {
		return links[0], nil
	}
}
