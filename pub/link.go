package pub

import (
	"net/url"
	"errors"
)

type Link struct {
	Object
}

func (l Link) Kind() string {
	return "Link"
}
func (l Link) Category() string {
	return "link"
}

func (l Link) Supertype() (string, error) {
	mediaType, err := l.GetMediaType("mediaType")
	if err != nil { return "", err }
	return mediaType.Supertype, nil
}

func (l Link) Subtype() (string, error) {
	mediaType, err := l.GetMediaType("mediaType")
	if err != nil { return "", err }
	return mediaType.Subtype, nil
}

func (l Link) URL() (*url.URL, error) {
	return l.GetURL("href")
}

func (l Link) Alt() (string, error) {
	alt, err := l.GetString("name")
	if alt == "" || err != nil {
		alt, err = l.GetString("href")
		if err != nil { return "", err }
	}
	return alt, nil
}

func (l Link) rating() uint64 {
	height, err := l.GetNumber("height")
	if err != nil { height = 1 }
	width, err := l.GetNumber("width")
	if err != nil { width = 1 }
	return height * width
}

func (l Link) String(width int) (string, error) {
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

func (l Link) Preview() (string, error) {
	return "todo", nil
}

func SelectBestLink(links []Link, supertype string) (Link, error) {
	if len(links) == 0 {
		return Link{}, errors.New("can't select best link of type " + supertype + "/* from an empty list")
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
		return Link{}, errors.New("can't select first Link from an empty list of links")
	} else {
		return links[0], nil
	}
}