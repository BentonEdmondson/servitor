package kinds

import (
	"net/url"
)

type Link Dict

// one of these should be omitted so
// Link isn't Content
func (l Link) Kind() (string, error) {
	return "link", nil
}
func (l Link) Category() string {
	return "link"
}

func (l Link) MediaType() (string, error) {
	return Get[string](l, "mediaType")
}

func (l Link) URL() (*url.URL, error) {
	return GetURL(l, "href")
}

func (l Link) Alt() (string, error) {
	return Get[string](l, "name")
}

func (l Link) Identifier() (*url.URL, error) {
	return nil, nil
}

// TODO: update of course to be nice markup of some sort
func (l Link) String() (string, error) {
	if url, err := l.URL(); err == nil {
		return url.String(), nil
	} else {
		return "", err
	}
}

// guide:
// Audio, Image, Video
// filter for ones with audio/, image/, video/
// as mime type, tiebreaker is resolution
// otherwise just take what you can get
// Article, Note, Page, Document
// probably honestly just take the first one

// probably provide the priorities as lists
// then write a function that looks up the list

// var priorities = map[string][]string{
// 	"image": []string{""}
// }

// given a Post, find the best link
// func GetLink(p Post) (Link, error) {
// 	kind, err := p.Kind()
// 	if err != nil {
// 		return nil, err
// 	}
// 	switch kind {
// 	// case "audio":
// 	// 	fallthrough
// 	// case "image":
// 	// 	fallthrough
// 	// case "video":
// 	// 	return GetBestLink(p)
// 	case "article":
// 		fallthrough
// 	case "document":
// 		fallthrough
// 	case "note":
// 		fallthrough
// 	case "page":
// 		return GetFirstLink(p)
// 	default:
// 		return nil, errors.New("Link extraction is not supported for type " + kind)
// 	}
// }

// pulls the link with the mime type that
// matches the Kind of the post, used for
// image, audio, video

// the reason this can't use GetContent is because GetContent
// treats strings as URLs used to find the end object,
// whereas in this context strings are URLs that are the href
// being the endpoint the Link represents
// func GetBestLink(p Post) (Link, error) {

// }

// pulls the first link
// func GetFirstLink(p Post) (Link, error) {
// 	values, err := GetList(p, "url")
// 	if err != nil {
// 		return nil, err
// 	}
	
// 	var individual any

// 	if len(values) == 0 {
// 		return nil, errors.New("Link is an empty list on the post")
// 	} else {
// 		individual = values[0]
// 	}

// 	switch narrowed := individual.(type) {
// 	case string:
// 		// here I should build the link out of the outer object
// 		return Link{"type": "Link", "href": narrowed}, nil
// 	case Dict:
// 		return Construct(narrowed)
// 	default:
// 		return nil, errors.New("The first URL entry on the post is a non-string, non-object. What?")
// 	}

// }

//
// GetLinks(p Post)
// similar to GetContent, but treats strings
// as Link.href, not as a reference to an object
// that should be fulfilled
// so whereas GetContent uses networking, GetLink
// does not


// GetBestLink - uses mime types/resolutions to determine best link
// of a list of Links

