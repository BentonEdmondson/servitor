package pub

import (
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"mimicry/ansi"
	"mimicry/client"
	"mimicry/mime"
	"mimicry/object"
	"mimicry/render"
	"mimicry/style"
	"net/url"
	"strings"
	"time"
)

type Actor struct {
	kind      string
	name      string
	nameErr   error
	handle    string
	handleErr error

	id *url.URL

	bio          string
	bioErr       error
	mediaType    *mime.MediaType
	mediaTypeErr error

	joined    time.Time
	joinedErr error

	pfp       *Link
	pfpErr    error
	banner    *Link
	bannerErr error

	posts    *Collection
	postsErr error
}

func NewActor(input any, source *url.URL) (*Actor, error) {
	o, id, err := client.FetchUnknown(input, source)
	if err != nil {
		return nil, err
	}
	return NewActorFromObject(o, id)
}

func NewActorFromObject(o object.Object, id *url.URL) (*Actor, error) {
	a := &Actor{}
	a.id = id
	var err error
	if a.kind, err = o.GetString("type"); err != nil {
		return nil, err
	}

	if !slices.Contains([]string{
		"Application", "Group", "Organization", "Person", "Service",
	}, a.kind) {
		return nil, fmt.Errorf("%w: %s is not an Actor", ErrWrongType, a.kind)
	}

	a.name, a.nameErr = o.GetNatural("name", "en")
	a.handle, a.handleErr = o.GetString("preferredUsername")
	a.bio, a.bioErr = o.GetNatural("summary", "en")
	if a.bio == "" {
		a.bioErr = object.ErrKeyNotPresent
	}
	a.mediaType, a.mediaTypeErr = o.GetMediaType("mediaType")
	a.joined, a.joinedErr = o.GetTime("published")

	a.pfp, a.pfpErr = getBestLink(o, "icon", "image")
	a.banner, a.bannerErr = getBestLink(o, "image", "image")

	a.posts, a.postsErr = getCollection(o, "outbox", a.id)
	return a, nil
}

func (a *Actor) Kind() string {
	return a.kind
}

func (a *Actor) Parents(quantity uint) ([]Tangible, Tangible) {
	return []Tangible{}, nil
}

func (a *Actor) Children() Container {
	/* the if is necessary because my understanding is
	   the first nil is a (*Collection)(nil) whereas
	   the second is (Container)(nil) */
	if a.posts == nil {
		return nil
	} else {
		return a.posts
	}
}

// TODO: here is where I'd put forgery errors in
func (a *Actor) Name() string {
	var output string
	if a.nameErr == nil {
		output = a.name
	} else if !errors.Is(a.nameErr, object.ErrKeyNotPresent) {
		output = style.Problem(a.nameErr)
	}

	if a.id != nil && !errors.Is(a.handleErr, object.ErrKeyNotPresent) {
		if output != "" {
			output += " "
		}
		if a.handleErr != nil {
			output += style.Problem(a.handleErr)
		} else {
			output += style.Italic("@" + a.handle + "@" + a.id.Host)
		}
	}

	if a.kind != "Person" {
		if output != "" {
			output += " "
		}
		output += "(" + strings.ToLower(a.kind) + ")"
	} else if output == "" {
		output = strings.ToLower(a.kind)
	}

	return style.Color(output)
}

func (a *Actor) header(width int) string {
	output := a.Name()

	if errors.Is(a.joinedErr, object.ErrKeyNotPresent) {
		// omit it
	} else if a.joinedErr != nil {
		output += "\njoined " + style.Problem(a.joinedErr)
	} else {
		output += "\njoined " + style.Color(a.joined.Format("2 Jan 2006"))
	}

	return ansi.Wrap(output, width)
}

func (a *Actor) center(width int) (string, bool) {
	if errors.Is(a.bioErr, object.ErrKeyNotPresent) {
		return "", false
	}
	if a.bioErr != nil {
		return ansi.Wrap(style.Problem(a.bioErr), width), true
	}

	mediaType := a.mediaType
	if errors.Is(a.mediaTypeErr, object.ErrKeyNotPresent) {
		mediaType = mime.Default()
	} else if a.mediaTypeErr != nil {
		return ansi.Wrap(style.Problem(a.mediaTypeErr), width), true
	}

	rendered, err := render.Render(a.bio, mediaType.Essence, width)
	if err != nil {
		return style.Problem(err), true
	}
	return rendered, true
}

func (a *Actor) footer(width int) (string, bool) {
	if errors.Is(a.postsErr, object.ErrKeyNotPresent) {
		return style.Problem(a.postsErr), true
	} else if a.postsErr != nil {
		return "", false
	} else if quantity, err := a.posts.Size(); errors.Is(err, object.ErrKeyNotPresent) {
		return "", false
	} else if err != nil {
		return style.Problem(err), true
	} else if quantity == 1 {
		return style.Color(fmt.Sprintf("%d post", quantity)), true
	} else {
		return style.Color(fmt.Sprintf("%d posts", quantity)), true
	}
}

func (a *Actor) String(width int) string {
	output := a.header(width)

	if body, present := a.center(width - 4); present {
		output += "\n\n" + ansi.Indent(body, "  ", true) + "\n"
	}

	if footer, present := a.footer(width); present {
		output += "\n" + footer
	}

	return output
}

func (a *Actor) Preview(width int) string {
	output := a.header(width)

	// TODO this needs to be truncated
	if body, present := a.center(width); present {
		output += "\n" + ansi.Snip(body, width, 4, style.Color("\u2026"))
	}

	if footer, present := a.footer(width); present {
		output += "\n" + footer
	}

	return output
}

func (a *Actor) Timestamp() time.Time {
	if a.joinedErr != nil {
		return time.Time{}
	} else {
		return a.joined
	}
}
