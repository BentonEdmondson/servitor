package kinds

import (
	"strings"
	"net/url"
	"mimicry/shared"
	"mimicry/style"
	"fmt"
)

type Actor map[string]any

func (a Actor) Kind() (string, error) {
	kind, err := shared.Get[string](a, "type")
	return strings.ToLower(kind), err
}

func (a Actor) Name() (string, error) {
	name, err := shared.GetNatural(a, "name", "en")
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
	return fmt.Sprintf("%s (%s)", name, id.Hostname()), nil
}

func (a Actor) Category() string {
	return "actor"
}

func (a Actor) Identifier() (*url.URL, error) {
	return shared.GetURL(a, "id")
}

func (a Actor) Bio() (string, error) {
	bio, err := shared.GetNatural(a, "summary", "en")
	return strings.TrimSpace(bio), err
}

func (a Actor) String() string {
	output := ""

	name, err := a.InlineName()
	if err == nil {
		output += style.Bold(name)
	}
	kind, err := a.Kind()
	if err == nil {
		output += " "
		output += kind
	}
	bio, err := a.Bio()
	if err == nil {
		output += "\n"
		output += bio
	}
	return output
}