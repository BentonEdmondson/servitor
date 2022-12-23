package kinds

import (
	"net/url"
	"strings"
	"time"
	"mimicry/style"
	"fmt"
	"errors"
)

type Post Dict

// TODO: make the Post references *Post because why not

func (p Post) Kind() (string, error) {
	kind, err := Get[string](p, "type")
	return strings.ToLower(kind), err
}

func (p Post) Title() (string, error) {
	title, err := GetNatural(p, "name", "en")
	return strings.TrimSpace(title), err
}

func (p Post) Body() (string, error) {
	body, err := GetNatural(p, "content", "en")
	return strings.TrimSpace(body), err
}

func (p Post) BodyPreview() (string, error) {
	body, err := p.Body()
	// probably should convert to runes and just work with that
	if len(body) > 280*2 { // this is a bug because len counts bytes whereas later I work based on runes
		return fmt.Sprintf("%s…", string([]rune(body)[:280])), err
	} else {
		return body, err
	}
}

func (p Post) Identifier() (*url.URL, error) {
	return GetURL(p, "id")
}

func (p Post) Created() (time.Time, error) {
	return GetTime(p, "published")
}

func (p Post) Updated() (time.Time, error) {
	return GetTime(p, "updated")
}

func (p Post) Category() string {
	return "post"
}

func (p Post) Creators() ([]Actor, error) {
	return GetContent[Actor](p, "attributedTo")
}

func (p Post) Attachments() ([]Link, error) {
	return GetAsLinks(p, "attachment")
}

// func (p Post) bestLink() (Link, error) {

// }

func (p Post) Link() (Link, error) {
	kind, err := p.Kind()
	if err != nil {
		return nil, err
	}
	switch kind {
	// case "audio", "image", "video":
	// 	return GetBestLink(p)
	case "article", "document", "note", "page":
		if links, err := GetLinks(p, "url"); err != nil {
			return nil, err
		} else if len(links) == 0 {
			return nil, err
		} else {
			return links[0], nil
		}
	default:
		return nil, errors.New("Link extraction is not supported for type " + kind)
	}
}

func (p Post) String() (string, error) {
	output := ""

	if title, err := p.Title(); err == nil {
		output += style.Bold(title)
		output += "\n"
	}


	if body, err := p.BodyPreview(); err == nil {
		output += body
		output += "\n"
	}

	if created, err := p.Created(); err == nil {
		output += time.Now().Sub(created).String()
	}

	if creators, err := p.Creators(); err == nil {
		output += " "
		for _, creator := range creators {
			if name, err := creator.InlineName(); err == nil {
				output += style.Bold(name) + ", "
			}
		}
	}

	if link, err := p.Link(); err == nil {
		if linkStr, err := link.String(); err == nil {
			output += "\n"
			output += linkStr
		}
	}

	if attachments, err := p.Attachments(); err == nil {
		output += "\nAttachments:\n"
		for _, attachment := range attachments {
			if attachmentStr, err := attachment.String(); err == nil {
				output += attachmentStr + "\n"
			} else {
				continue
			}
		}
	}

	return strings.TrimSpace(output), nil
}
