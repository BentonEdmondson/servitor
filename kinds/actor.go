package kinds

import (
	"strings"
	"net/url"
	"mimicry/style"
)

type Actor struct {
	Object
}

func (a Actor) Kind() string {
	kind, err := a.GetString("type")
	if err != nil {
		panic(err)
	}
	return strings.ToLower(kind)
}

func (a Actor) Name() (string, error) {
	if a.Has("preferredUsername") && !a.HasNatural("name") {
		name, err := a.GetString("preferredUsername")
		if err != nil { return "", err }
		return "@" + name, nil
	}
	return a.GetNatural("name", "en")
}

func (a Actor) InlineName() (string, error) {
	name, err := a.Name()
	if err != nil {
		return "", err
	}
	kind := a.Kind()
	var suffix string
	id, err := a.Identifier()
	if err == nil {
		if kind == "person" {
			suffix = "(" + id.Hostname() + ")"
		} else {
			suffix = "(" + id.Hostname() + ", " + kind + ")"
		}
	}
	return name + " " + suffix, nil
}

func (a Actor) Category() string {
	return "actor"
}

func (a Actor) Identifier() (*url.URL, error) {
	return a.GetURL("id")
}

func (a Actor) Bio() (string, error) {
	return a.Render("summary", "en", "mediaType", 80)
}

func (a Actor) String(width int) (string, error) {
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

func (a Actor) Preview() (string, error) {
	return "todo", nil
}