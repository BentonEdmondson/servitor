package kinds

import (
	"net/url"
	"strings"
	"time"
	"mimicry/style"
	"fmt"
	"errors"
	"mimicry/render"
)

type Post Dict

// TODO: go through and remove all the trims, they
// make things less predictable
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
	mediaType, err := Get[string](p, "mediaType")
	if err != nil {
		mediaType = "text/html"
	}
	return render.Render(body, mediaType)
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
	return GetLinksLenient(p, "attachment")
}

// func (p Post) bestLink() (Link, error) {

// }

func (p Post) Link() (Link, error) {
	kind, err := p.Kind()
	if err != nil {
		return nil, err
	}

	links, err := GetLinksStrict(p, "url")
	if err != nil {
		return nil, err
	}

	switch kind {
	case "audio", "image", "video":
		return SelectBestLink(links, kind)
	case "article", "document", "note", "page":
		return SelectFirstLink(links)
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


	if body, err := p.Body(); err == nil {
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
