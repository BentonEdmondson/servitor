package kinds

import (
	"net/url"
	"strings"
	"time"
	"mimicry/shared"
	"mimicry/style"
	"mimicry/request"
	"fmt"
)

type Post map[string]any

// TODO: make the Post references *Post because why not

func (p Post) Kind() (string, error) {
	kind, err := shared.Get[string](p, "type")
	return strings.ToLower(kind), err
}

func (p Post) Title() (string, error) {
	title, err := shared.GetNatural(p, "name", "en")
	return strings.TrimSpace(title), err
}

func (p Post) Body() (string, error) {
	body, err := shared.GetNatural(p, "content", "en")
	return strings.TrimSpace(body), err
}

func (p Post) BodyPreview() (string, error) {
	body, err := p.Body()
	return fmt.Sprintf("%s…", string([]rune(body)[:280])), err
}

func (p Post) Identifier() (*url.URL, error) {
	return shared.GetURL(p, "id")
}

func (p Post) Created() (time.Time, error) {
	return shared.GetTime(p, "published")
}

func (p Post) Updated() (time.Time, error) {
	return shared.GetTime(p, "updated")
}

func (p Post) Category() string {
	return "post"
}

func (p Post) Creators() []Actor {
	// TODO: this line needs an existence check
	attributedTo, ok := p["attributedTo"]
	if !ok {
		return []Actor{}
	}

	// if not an array, make it an array
	attributions := []any{}
	if attributedToList, isList := attributedTo.([]any); isList {
		attributions = attributedToList
	} else {
		attributions = []any{attributedTo}
	}

	output := []Actor{}

	for _, el := range attributions {
		switch narrowed := el.(type) {
		case shared.JSON:
			source, err := p.Identifier()
			if err != nil { continue }
			resolved, err := Create(narrowed, source)
			if err != nil { continue }
			actor, isActor := resolved.(Actor)
			if !isActor { continue }
			output = append(output, actor)
		case string:
			url, err := url.Parse(narrowed)
			if err != nil { continue }
			response, err := request.Fetch(url)
			if err != nil { continue }
			// this step will be implicit after merge
			structured, err := Create(response, url)
			if err != nil { continue }
			actor, isActor := structured.(Actor)
			if !isActor { continue }
			output = append(output, actor)
		default: continue
		}
	}

	return output
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

	if creators := p.Creators(); len(creators) != 0 {
		output += " "
		for _, creator := range creators {
			if name, err := creator.InlineName(); err == nil {
				output += style.Bold(name) + ", "
			}
		}
	}

	return strings.TrimSpace(output)
}
