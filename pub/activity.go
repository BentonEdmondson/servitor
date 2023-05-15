package pub

import (
	"net/url"
	"mimicry/object"
	"mimicry/client"
	"fmt"
	"golang.org/x/exp/slices"
	"mimicry/ansi"
	"mimicry/style"
)

type Activity struct {
	kind string
	id *url.URL

	actor *Actor; actorErr error
	target Tangible
}

func NewActivity(input any, source *url.URL) (*Activity, error) {
	a := &Activity{}
	var err error; var o object.Object
	o, a.id, err = client.FetchUnknown(input, source)
	if err != nil { return nil, err }
	if a.kind, err = o.GetString("type"); err != nil {
		return nil, err
	}

	if !slices.Contains([]string{
		"Create", "Announce", "Dislike", "Like",
	}, a.kind) {
		return nil, fmt.Errorf("%w: %s is not an Activity", ErrWrongType, a.kind)
	}

	// TODO: parallelize
	a.actor, a.actorErr = getActor(o, "actor", a.id)
	a.target = getPostOrActor(o, "object", a.id)

	return a, nil
}

func (a *Activity) Kind() string {
	return a.kind
}

func (a *Activity) header(width int) string {
	if a.kind == "Create" {
		return ""
	}

	var output string
	if a.actorErr != nil {
		output += style.Problem(a.actorErr)
	} else {
		output += a.actor.Name()
	}

	output += " "

	switch a.kind {
	case "Announce":
		output += style.Color("retweeted")
	case "Like":
		output += style.Color("upvoted")
	case "Dislike":
		output += style.Color("downvoted")
	default:
		panic("encountered unrecognized Actor type: " + a.kind)
	}

	output += ":\n"

	return ansi.Wrap(output, width)
}

func (a *Activity) String(width int) string {
	output := a.header(width)

	output += a.target.String(width)
	return output
}

func (a *Activity) Preview(width int) string {
	output := a.header(width)

	output += a.target.Preview(width)
	return output
}

func (a *Activity) Children(quantity uint) ([]Tangible, Container, uint) {
	return a.target.Children(quantity)
}

func (a *Activity) Parents(quantity uint) []Tangible {
	return a.target.Parents(quantity)
}
