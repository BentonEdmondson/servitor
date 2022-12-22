package kinds

import (
	"net/url"
	"strings"
	"time"
	"mimicry/style"
	"fmt"
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
	return fmt.Sprintf("%s…", string([]rune(body)[:280])), err
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

func (p Post) String() string {
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

	return strings.TrimSpace(output)
}
