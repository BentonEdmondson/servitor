package kinds

import (
	"strings"
	"net/url"
	"mimicry/style"
	"fmt"
	"mimicry/render"
)

type Actor Dict

func (a Actor) Kind() (string, error) {
	kind, err := Get[string](a, "type")
	return strings.ToLower(kind), err
}

func (a Actor) Name() (string, error) {
	name, err := GetNatural(a, "name", "en")
	return strings.TrimSpace(name), err
}

func (a Actor) InlineName() (string, error) {
	name, err := a.Name()
	if err != nil {
		return "", err
	}
	id, err := a.Identifier()
	if err != nil {
		return "", err
	}
	kind, err := a.Kind()
	if err != nil {
		return "", err
	}
	// if kind != "person" {
	// 	return fmt.Sprintf("%s (%s, %s)", name, id.Hostname(), kind), nil
	// }
	// return fmt.Sprintf("%s (%s)", name, id.Hostname()), nil
	return fmt.Sprintf("%s (%s, %s)", name, id.Hostname(), kind), nil
}

func (a Actor) Category() string {
	return "actor"
}

func (a Actor) Identifier() (*url.URL, error) {
	return GetURL(a, "id")
}

func (a Actor) Bio() (string, error) {
	body, err := GetNatural(a, "summary", "en")
	mediaType, err := Get[string](a, "mediaType")
	if err != nil {
		mediaType = "text/html"
	}
	return render.Render(body, mediaType, 80)
}

func (a Actor) String() (string, error) {
	output := ""

	name, err := a.InlineName()
	if err == nil {
		output += style.Bold(name)
	}
	bio, err := a.Bio()
	if err == nil {
		output += "\n"
		output += bio
	}
	return output, nil
}