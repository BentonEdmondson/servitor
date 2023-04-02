package pub

import (
	"net/url"
	"strings"
	"time"
	"mimicry/style"
	"mimicry/ansi"
)

type Post struct {
	Object
}

func (p Post) Kind() (string) {
	kind, err := p.GetString("type")
	if err != nil {
		panic(err)
	}
	return strings.ToLower(kind)
}

func (p Post) Title() (string, error) {
	return p.GetNatural("name", "en")
}

func (p Post) Body(width int) (string, error) {
	return p.Render("content", "en", "mediaType", width)
}

func (p Post) Identifier() (*url.URL, error) {
	return p.GetURL("id")
}

func (p Post) Created() (time.Time, error) {
	return p.GetTime("published")
}

func (p Post) Edited() (time.Time, error) {
	return p.GetTime("updated")
}

func (p Post) Category() string {
	return "post"
}

func (p Post) Creators() ([]Actor, error) {
	return p.GetActors("attributedTo")
}

func (p Post) Recipients() ([]Actor, error) {
	return p.GetActors("to")
}

func (p Post) Attachments() ([]Link, error) {
	return p.GetLinks("attachment")
}

func (p Post) Comments() (Collection, error) {
	if p.Has("comments") && !p.Has("replies") {
		return p.GetCollection("comments")
	}
	return p.GetCollection("replies")
}

func (p Post) Link() (Link, error) {
	values, err := p.GetList("url")
	if err != nil {
		return Link{}, err
	}
	
	links := make([]Link, 0, len(values))

	for _, el := range values {
		switch narrowed := el.(type) {
		case string:
			link := Link{Object{
				"type": "Link",
				"href": narrowed,
			}}
			if name, err := p.GetNatural("name", "en"); err == nil {
				link.Object["name"] = name
			}
			if !p.HasNatural("content") {
				if mediaType, err := p.GetString("mediaType"); err == nil {
					link.Object["mediaType"] = mediaType
				}
			}
			links = append(links, link)
		case Object:
			source, _ := p.GetURL("id")
			item, err := Construct(narrowed, source)
			if err != nil { continue }
			if asLink, isLink := item.(Link); isLink {
				links = append(links, asLink)
			}
		}
	}

	kind := p.Kind()
	switch kind {
	case "audio", "image", "video":
		return SelectBestLink(links, kind)
	default:
		return SelectFirstLink(links)
	}
}

func (p Post) header(width int) (string, error) {
	output := ""

	if title, err := p.Title(); err == nil {
		output += style.Bold(title) + "\n"
	}

	output += style.Color(p.Kind())

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

	return output, nil
}
