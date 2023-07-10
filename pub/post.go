package pub

import (
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"servitor/ansi"
	"servitor/client"
	"servitor/mime"
	"servitor/object"
	"servitor/style"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Post struct {
	kind string
	id   *url.URL

	title      string
	titleErr   error
	body       object.Markup
	bodyLinks  []string
	bodyErr    error
	media      *Link
	mediaErr   error
	created    time.Time
	createdErr error
	edited     time.Time
	editedErr  error
	parentObject     object.Object
	parentIdentifier *url.URL
	parentErr  error

	// just as body dies completely if members die,
	// attachments dies completely if any member dies
	attachments    []*Link
	attachmentsErr error

	creators    []Tangible
	recipients  []Tangible
	comments    *Collection
	commentsErr error
}

func NewPost(input any, source *url.URL) (*Post, error) {
	o, id, err := client.FetchUnknown(input, source)
	if err != nil {
		return nil, err
	}
	return NewPostFromObject(o, id)
}

func NewPostFromObject(o object.Object, id *url.URL) (*Post, error) {
	p := &Post{}
	p.id = id
	var err error
	if p.kind, err = o.GetString("type"); err != nil {
		return nil, err
	}

	if p.kind == "Tombstone" {
		return nil, errors.New("post was deleted")
	}

	if !slices.Contains([]string{
		"Article", "Audio", "Document", "Image", "Note", "Page", "Video",
	}, p.kind) {
		return nil, fmt.Errorf("%w: %s is not a Post", ErrWrongType, p.kind)
	}

	p.title, p.titleErr = o.GetString("name")
	p.body, p.bodyLinks, p.bodyErr = o.GetMarkup("content", "mediaType")
	p.created, p.createdErr = o.GetTime("published")
	p.edited, p.editedErr = o.GetTime("updated")
	p.parentObject, p.parentIdentifier, p.parentErr = getAndFetchUnkown(o, "inReplyTo", p.id)

	if p.kind == "Audio" || p.kind == "Video" || p.kind == "Image" {
		p.media, p.mediaErr = getBestLinkShorthand(o, "url", strings.ToLower(p.kind))
	} else {
		p.media, p.mediaErr = getFirstLinkShorthand(o, "url")
	}

	var wg sync.WaitGroup
	wg.Add(4)
	go func() { p.creators = getActors(o, "attributedTo", p.id); wg.Done() }()
	go func() { p.recipients = getActors(o, "audience", p.id); wg.Done() }()
	go func() { p.attachments, p.attachmentsErr = getLinks(o, "attachment"); wg.Done() }()

	constructComment := func(input any, source *url.URL) Tangible {
		comment, err := NewPost(input, source)
		if err != nil {
			return NewFailure(err)
		}

		if id == nil {
			return NewFailure(errors.New("comment does not reference this parent (parent lacks an identifier)"))
		}

		if comment.ParentIdentifier() == nil || comment.ParentIdentifier().String() != id.String() {
			return NewFailure(errors.New("comment does not reference this parent"))
		}

		return comment
	}

	go func() {
		p.comments, p.commentsErr = getCollection(o, "replies", p.id, constructComment)
		if errors.Is(p.commentsErr, object.ErrKeyNotPresent) {
			p.comments, p.commentsErr = getCollection(o, "comments", p.id, constructComment)
		}
		wg.Done()
	}()
	wg.Wait()

	/* Ensure that creators come from the same host as the post itself */
	for _, creator := range p.creators {
		if asActor, isActor := creator.(*Actor); isActor {
			if asActor.Identifier() == nil && id == nil {
				continue
			}

			if (asActor.Identifier() == nil || id == nil) || asActor.Identifier().Host != id.Host {
				return nil, errors.New("post contains forged creators")
			}
		}
		/* These are necessarily Failure types, so don't need to be checked */
	}

	return p, nil
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
		if errors.Is(p.parentErr, object.ErrKeyNotPresent) {
			return []Tangible{}, nil
		}
		return []Tangible{}, p
	}
	if errors.Is(p.parentErr, object.ErrKeyNotPresent) {
		return []Tangible{}, nil
	}
	if p.parentErr != nil {
		return []Tangible{NewFailure(p.parentErr)}, nil
	}
	parent, err := NewPostFromObject(p.parentObject, p.parentIdentifier)
	if err != nil {
		return []Tangible{NewFailure(err)}, nil
	}
	if quantity == 1 {
		return []Tangible{parent}, parent
	}
	parentParents, parentFrontier := parent.Parents(quantity - 1)
	return append([]Tangible{parent}, parentParents...), parentFrontier
}

func (p *Post) ParentIdentifier() *url.URL {
	if p.parentErr != nil {
		return nil
	}
	return p.parentIdentifier
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

	/* TODO: forgery checking is needed here; verify that the id of the post
	   and id of the creators match */
	if len(p.creators) > 0 {
		output += " by "
		for i, creator := range p.creators {
			output += style.Color(creator.Name())
			if i != len(p.creators)-1 {
				output += ", "
			}
		}
	}
	if len(p.recipients) > 0 {
		output += " to "
		for i, recipient := range p.recipients {
			output += style.Color(recipient.Name())
			if i != len(p.recipients)-1 {
				output += ", "
			}
		}
	}

	if p.createdErr != nil && !errors.Is(p.createdErr, object.ErrKeyNotPresent) {
		output += " at " + style.Problem(p.createdErr)
	} else {
		output += " • " + style.Color(ago(p.created))
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

	rendered := p.body.Render(width)
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

	// TODO: don't think this is good, rework it
	output := ""
	for i, attachment := range p.attachments {
		if output != "" {
			output += "\n"
		}
		alt, err := attachment.Alt()
		if err != nil {
			output += style.Problem(err)
			continue
		}
		output += style.LinkBlock(ansi.Wrap(alt, width-2), len(p.bodyLinks)+i+1)
	}
	return output, true
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

	body, bodyPresent := p.center(width)
	if bodyPresent {
		output += "\n" + body
	}

	if attachments, present := p.supplement(width); present {
		if bodyPresent {
			output += "\n"
		}
		output += "\n" + attachments
	}

	return ansi.Snip(output, width, 4, style.Color("\u2026"))
}

func (p *Post) Timestamp() time.Time {
	if p.createdErr != nil {
		return time.Time{}
	} else {
		return p.created
	}
}

func (p *Post) Name() string {
	if p.titleErr != nil {
		return style.Problem(p.titleErr)
	}
	return p.title
}

func (p *Post) Creators() []Tangible {
	return p.creators
}

func (p *Post) Recipients() []Tangible {
	return p.recipients
}

func (p *Post) Media() (string, *mime.MediaType, bool) {
	if p.mediaErr != nil {
		return "", nil, false
	}

	if p.kind == "Audio" || p.kind == "Video" || p.kind == "Image" {
		return p.media.SelectWithDefaultMediaType(mime.UnknownSubtype(strings.ToLower(p.kind)))
	}

	return p.media.Select()
}

func (p *Post) SelectLink(input int) (string, *mime.MediaType, bool) {
	input -= 1
	if len(p.bodyLinks) > input {
		return p.bodyLinks[input], mime.Unknown(), true
	}
	nextIndex := input - len(p.bodyLinks)
	if len(p.attachments) > nextIndex {
		return p.attachments[nextIndex].Select()
	}
	return "", nil, false
}
