package kinds

import (
	"net/url"
	"strings"
	"errors"
)

type Link Dict

// one of these should be omitted so
// Link isn't Content
func (l Link) Kind() (string, error) {
	return "link", nil
}
func (l Link) Category() string {
	return "link"
}

func (l Link) Supertype() (string, error) {
	mediaType, err := Get[string](l, "mediaType")
	return strings.Split(mediaType, "/")[0], err
}

func (l Link) Subtype() (string, error) {
	if mediaType, err := Get[string](l, "mediaType"); err != nil {
		return "", err
	} else if split := strings.Split(mediaType, "/"); len(split) < 2 {
		return "", errors.New("Media type " + mediaType + " lacks a subtype")
	} else {
		return split[1], nil
	}
}

func (l Link) URL() (*url.URL, error) {
	return GetURL(l, "href")
}

func (l Link) Alt() (string, error) {
	alt, err := Get[string](l, "name")
	return strings.TrimSpace(alt), err
}

func (l Link) Identifier() (*url.URL, error) {
	return nil, nil
}

// used for link prioritization, roughly
// related to resolution
func (l Link) rating() int {
	height, err := Get[int](l, "height")
	if err != nil { height = 1 }
	width, err := Get[int](l, "width")
	if err != nil { width = 1 }
	return height * width
}

// TODO: update of course to be nice markup of some sort
func (l Link) String() (string, error) {
	output := ""

	if alt, err := l.Alt(); err == nil {
		output += alt
	} else if url, err := l.URL(); err == nil {
		output += url.String()
	}

	if Subtype, err := l.Subtype(); err == nil {
		output += " (" + Subtype + ")"
	}

	return output, nil
}

// TODO: must test when list only has 1 link (probably works)
func SelectBestLink(links []Link, supertype string) (Link, error) {
	if len(links) == 0 {
		return nil, errors.New("Can't select best link of type " + supertype + "/* from an empty list")
	}

	bestLink := links[0]

	for _, thisLink := range links[1:] {
		var bestLinkSupertypeMatches bool
		if bestLinkSupertype, err := bestLink.Supertype(); err != nil {
			bestLinkSupertypeMatches = false
		} else {
			bestLinkSupertypeMatches = bestLinkSupertype == supertype
		}

		var thisLinkSuperTypeMatches bool
		if thisLinkSupertype, err := thisLink.Supertype(); err != nil {
			thisLinkSuperTypeMatches = false
		} else {
			thisLinkSuperTypeMatches = thisLinkSupertype == supertype
		}

		if thisLinkSuperTypeMatches && !bestLinkSupertypeMatches {
			bestLink = thisLink
			continue
		} else if !thisLinkSuperTypeMatches && bestLinkSupertypeMatches {
			continue
		} else if thisLink.rating() > bestLink.rating() {
			bestLink = thisLink
			continue
		}
	}

	return bestLink, nil
}

func SelectFirstLink(links []Link) (Link, error) {
	if len(links) == 0 {
		return nil, errors.New("can't select first Link from an empty list of links")
	} else {
		return links[0], nil
	}
}