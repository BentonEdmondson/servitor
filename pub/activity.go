package pub

import (
	"net/url"
	"mimicry/object"
	"mimicry/client"
	"fmt"
	"golang.org/x/exp/slices"
	"mimicry/ansi"
	"mimicry/style"
	"sync"
)

type Activity struct {
	kind string
	id *url.URL

	actor *Actor; actorErr error
	target Tangible
}

func NewActivity(input any, source *url.URL) (*Activity, error) {
	o, id, err := client.FetchUnknown(input, source)
	if err != nil { return nil, err }
	return NewActivityFromObject(o, id)
}

func NewActivityFromObject(o object.Object, id *url.URL) (*Activity, error) {
	a := &Activity{}
	a.id = id
	var err error
	if a.kind, err = o.GetString("type"); err != nil {
		return nil, err
	}

	if !slices.Contains([]string{
		"Create", "Announce", "Dislike", "Like",
	}, a.kind) {
		return nil, fmt.Errorf("%w: %s is not an Activity", ErrWrongType, a.kind)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func () {a.actor, a.actorErr = getActor(o, "actor", a.id); wg.Done()}()
	go func() {a.target = getPostOrActor(o, "object", a.id); wg.Done()}()
	wg.Wait()

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
		output += "retweeted"
	case "Like":
		output += "upvoted"
	case "Dislike":
		output += "downvoted"
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

func (a *Activity) Children() Container {
	return a.target.Children()
}

func (a *Activity) Parents(quantity uint) ([]Tangible, Tangible) {
	return a.target.Parents(quantity)
}
