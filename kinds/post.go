package kinds

import (
	"net/url"
	"strings"
	"time"
	"mimicry/style"
	"mimicry/render"
	"mimicry/ansi"
)

type Post Dict

// TODO: go through and remove all the trims, they
// make things less predictable
// TODO: make the Post references *Post because why not

func (p Post) Raw() Dict {
	return p
}

func (p Post) Kind() (string, error) {
	kind, err := Get[string](p, "type")
	return strings.ToLower(kind), err
}

func (p Post) Title() (string, error) {
	title, err := GetNatural(p, "name", "en")
	return strings.TrimSpace(title), err
}

func (p Post) Body(width int) (string, error) {
	body, err := GetNatural(p, "content", "en")
	if err != nil {
		return "", err
	}
	mediaType, err := Get[string](p, "mediaType")
	if err != nil {
		mediaType = "text/html"
	}
	return render.Render(body, mediaType, width)
}

func (p Post) Identifier() (*url.URL, error) {
	return GetURL(p, "id")
}

func (p Post) Created() (time.Time, error) {
	return GetTime(p, "published")
}

// TODO: rename to edited
func (p Post) Updated() (time.Time, error) {
	return GetTime(p, "updated")
}

func (p Post) Category() string {
	return "post"
}

func (p Post) Creators() ([]Actor, error) {
	return GetContent[Actor](p, "attributedTo")
}

func (p Post) Recipients() ([]Actor, error) {
	return GetContent[Actor](p, "to")
}

func (p Post) Attachments() ([]Link, error) {
	return GetLinksLenient(p, "attachment")
}

func (p Post) Comments() (Collection, error) {
	replies, repliesErr := GetItem[Collection](p, "replies")
	if repliesErr != nil {
		comments, commentsErr := GetItem[Collection](p, "comments")
		if commentsErr != nil {
			return Collection{}, repliesErr
		}
		replies = comments
	}
	return replies, nil
}

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
	default: // "article", "document", "note", "page":
		return SelectFirstLink(links)
	}
}

func (p Post) header(width int) (string, error) {
	output := ""

	if title, err := p.Title(); err == nil {
		output += style.Bold(title) + "\n"
	}

	if kind, err := p.Kind(); err == nil {
		output += style.Color(kind)
	}

	if creators, err := p.Creators(); err == nil {
		names := []string{}
		for _, creator := range creators {
			if name, err := creator.InlineName(); err == nil {
				names = append(names, style.Link(name))
			}
		}
		if len(names) > 0 {
			output += " by " + strings.Join(names, ", ")
		}
	}

	if recipients, err := p.Recipients(); err == nil {
		names := []string{}
		for _, recipient := range recipients {
			if name, err := recipient.InlineName(); err == nil {
				names = append(names, style.Link(name))
			}
		}
		if len(names) > 0 {
			output += " to " + strings.Join(names, ", ")
		}
	}

	if created, err := p.Created(); err == nil {
		const timeFormat = "3:04 pm on 2 Jan 2006"
		output += " at " + style.Color(created.Format(timeFormat))
		// if edited, err := p.Updated(); err == nil {
		// 	output += " (edited at " + style.Color(edited.Format(timeFormat)) + ")"
		// }
	}

	return ansi.Wrap(output, width), nil
}

func (p Post) String(width int) (string, error) {
	output := ""

	if header, err := p.header(width - 4); err == nil {
		output += ansi.Indent(header, "  ", true)
		output += "\n\n"
	}

	if body, err := p.Body(width - 8); err == nil {
		output += ansi.Indent(body, "    ", true)
		output += "\n\n"
	}

	if attachments, err := p.Attachments(); err == nil {
		if len(attachments) > 0 {
			section := "Attachments:\n"
			names := []string{}
			for _, attachment := range attachments {
				if name, err := attachment.String(width); err == nil {
					names = append(names, style.Link(name))
				}
			}
			section += ansi.Indent(ansi.Wrap(strings.Join(names, "\n"), width - 4), "  ", true)
			section = ansi.Indent(ansi.Wrap(section, width - 2), "  ", true)
			output += section
			output += "\n"
		}
	}

	if comments, err := p.Comments(); err == nil {
		if size, err := comments.Size(); err == nil {
			output += ansi.Indent(ansi.Wrap("with " + style.Color(size + " comments"), width - 2), "  ", true)
			output += "\n\n"
		}
		if section, err := comments.String(width); err == nil {
			output += section + "\n"
		} else {
			return "", err
		}
	} else {
		return "", err
	}

	return output, nil
}

func (p Post) Preview() (string, error) {
	output := ""
	width := 100

	if header, err := p.header(width); err == nil {
		output += header
		output += "\n"
	}

	if body, err := p.Body(width); err == nil {
		output += ansi.Snip(body, width, 4, style.Color("\u2026"))
		output += "\n"
	}

	// TODO: there should probably be attachments here

	return output, nil
}
