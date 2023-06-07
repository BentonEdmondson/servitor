package pub

import (
	"net/url"
	"strings"
	"time"
	"mimicry/style"
	"mimicry/ansi"
	"mimicry/object"
	"errors"
	"mimicry/client"
	"fmt"
	"golang.org/x/exp/slices"
	"mimicry/mime"
	"mimicry/render"
	"sync"
)

type Post struct {
	kind string
	id *url.URL

	title string
	titleErr error
	body string
	bodyErr error
	mediaType *mime.MediaType
	mediaTypeErr error
	link *Link
	linkErr error
	created time.Time
	createdErr error
	edited time.Time
	editedErr error
	parent any
	parentErr error

	// just as body dies completely if members die,
	// attachments dies completely if any member dies
	attachments []*Link
	attachmentsErr error

	creators []TangibleWithName
	recipients []TangibleWithName
	comments *Collection
	commentsErr error
}

func NewPost(input any, source *url.URL) (*Post, error) {
	o, id, err := client.FetchUnknown(input, source)
	if err != nil { return nil, err }
	return NewPostFromObject(o, id)
}

func NewPostFromObject(o object.Object, id *url.URL) (*Post, error) {
	p := &Post{}
	p.id = id
	var err error
	if p.kind, err = o.GetString("type"); err != nil {
		return nil, err
	}

	// TODO: for Lemmy, may have to auto-unwrap Create into a Post
	if !slices.Contains([]string{
		"Article", "Audio", "Document", "Image", "Note", "Page", "Video",
	}, p.kind) {
		return nil, fmt.Errorf("%w: %s is not a Post", ErrWrongType, p.kind)
	}

	p.title, p.titleErr = o.GetNatural("name", "en")
	p.body, p.bodyErr = o.GetNatural("content", "en")
	p.mediaType, p.mediaTypeErr = o.GetMediaType("mediaType")
	p.created, p.createdErr = o.GetTime("published")
	p.edited, p.editedErr = o.GetTime("updated")
	p.parent, p.parentErr = o.GetAny("inReplyTo")
	
	if p.kind == "Image" || p.kind == "Audio" || p.kind == "Video" {
		p.link, p.linkErr = getBestLinkShorthand(o, "url", strings.ToLower(p.kind))
	} else {
		p.link, p.linkErr = getFirstLinkShorthand(o, "url")
	}

	var wg sync.WaitGroup
	wg.Add(4)
	go func() {p.creators = getActors(o, "attributedTo", p.id); wg.Done()}()
	go func() {p.recipients = getActors(o, "audience", p.id); wg.Done()}()
	go func() {p.attachments, p.attachmentsErr = getLinks(o, "attachment"); wg.Done()}()
	go func() {
		p.comments, p.commentsErr = getCollection(o, "replies", p.id)
		if errors.Is(p.commentsErr, object.ErrKeyNotPresent) {
			p.comments, p.commentsErr = getCollection(o, "comments", p.id)
		}
		wg.Done()
	}()
	wg.Wait()
	return p, nil
}

func (p *Post) Kind() (string) {
	return p.kind
}

func (p *Post) Children() Container {
	/* the if is necessary because my understanding is
	the first nil is a (*Collection)(nil) whereas
	the second is (Container)(nil) */
	if p.comments == nil {
		return nil
	} else {
		return p.comments
	}
}

func (p *Post) Parents(quantity uint) ([]Tangible, Tangible) {
	if quantity == 0 {
		panic("can't fetch 0 parents")
	}
	if errors.Is(p.parentErr, object.ErrKeyNotPresent) {
		return []Tangible{}, nil
	}
	if p.parentErr != nil {
		return []Tangible{NewFailure(p.parentErr)}, nil
	}
	fetchedParent, fetchedParentErr := NewPost(p.parent, p.id)
	if fetchedParentErr != nil {
		return []Tangible{NewFailure(fetchedParentErr)}, nil
	}
	if quantity == 1 {
		return []Tangible{fetchedParent}, fetchedParent
	}
	fetchedParentParents, fetchedParentFrontier := fetchedParent.Parents(quantity - 1)
	return append([]Tangible{fetchedParent}, fetchedParentParents...), fetchedParentFrontier
}

func (p *Post) header(width int) string {
	output := ""

	if p.titleErr == nil {
		output += style.Bold(p.title) + "\n"
	} else if !errors.Is(p.titleErr, object.ErrKeyNotPresent) {
		output += style.Problem(fmt.Errorf("failed to get title: %w", p.titleErr)) + "\n"
	}

	if errors.Is(p.parentErr, object.ErrKeyNotPresent) {
		output += style.Color(strings.ToLower(p.kind))
	} else {
		output += style.Color("comment")
	}

	if len(p.creators) > 0 {
		output += " by "
		for i, creator := range p.creators {
			output += style.Color(creator.Name())
			if i != len(p.creators) - 1 {
				output += ", "
			}
		}
	}
	if len(p.recipients) > 0 {
		output += " to "
		for i, recipient := range p.recipients {
			output += style.Color(recipient.Name())
			if i != len(p.recipients) - 1 {
				output += ", "
			}
		}
	}

	if p.createdErr != nil && !errors.Is(p.createdErr, object.ErrKeyNotPresent) {
		output += " at " + style.Problem(p.createdErr)
	} else {
		output += " at " + style.Color(time.Since(p.created).Round(time.Minute).String())
	}

	return ansi.Wrap(output, width)
}

func (p *Post) center(width int) (string, bool) {
	if errors.Is(p.bodyErr, object.ErrKeyNotPresent) {
		return "", false
	}
	if p.bodyErr != nil {
		return ansi.Wrap(style.Problem(p.bodyErr), width), true
	}

	mediaType := p.mediaType
	if errors.Is(p.mediaTypeErr, object.ErrKeyNotPresent) {
		mediaType = mime.Default()
	} else if p.mediaTypeErr != nil {
		return ansi.Wrap(style.Problem(p.mediaTypeErr), width), true
	}

	rendered, err := render.Render(p.body, mediaType.Essence, width)
	if err != nil {
		return style.Problem(err), true
	}
	return rendered, true
}

func (p *Post) supplement(width int) (string, bool) {
	if errors.Is(p.attachmentsErr, object.ErrKeyNotPresent) {
		return "", false
	}
	if p.attachmentsErr != nil {
		return ansi.Wrap(style.Problem(fmt.Errorf("failed to load attachments: %w", p.attachmentsErr)), width), true
	}
	if len(p.attachments) == 0 {
		return "", false
	}

	output := ""
	for _, attachment := range p.attachments {
		if output != "" { output += "\n" }
		alt, err := attachment.Alt()
		if err != nil {
			output += style.Problem(err)
			continue
		}
		output += style.LinkBlock(alt)
	}
	return ansi.Wrap(output, width), true
}

func (p *Post) footer(width int) string {
	if errors.Is(p.commentsErr, object.ErrKeyNotPresent) {
		return style.Color("comments disabled")
	} else if p.commentsErr != nil {
		return style.Color("comments enabled")
	} else if quantity, err := p.comments.Size(); errors.Is(err, object.ErrKeyNotPresent) {
		return style.Color("comments enabled")
	} else if err != nil {
		return style.Problem(err)
	} else if quantity == 1 {
		return style.Color(fmt.Sprintf("%d comment", quantity))
	} else {
		return style.Color(fmt.Sprintf("%d comments", quantity))
	}
}

func (p Post) String(width int) string {
	output := p.header(width)

	if body, present := p.center(width - 4); present {
		output += "\n\n" + ansi.Indent(body, "  ", true)
	}

	if attachments, present := p.supplement(width - 4); present {
		output += "\n\n" + ansi.Indent(attachments, "  ", true)
	}
	
	output += "\n\n" + p.footer(width)

	return output
}

func (p *Post) Preview(width int) string {
	output := p.header(width)

	if body, present := p.center(width); present {
		if attachments, present := p.supplement(width); present {
			output += "\n" + ansi.Snip(body + "\n" + attachments, width, 4, style.Color("\u2026"))
		} else {
			output += "\n" + ansi.Snip(body, width, 4, style.Color("\u2026"))
		}
	}

	return output
}

func (p *Post) Timestamp() time.Time {
	if p.createdErr != nil {
		return time.Time{}
	} else {
		return p.created
	}
}
