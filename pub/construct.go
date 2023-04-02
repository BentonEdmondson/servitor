package pub

import (
	"errors"
	"net/url"
	"strings"
	"mimicry/jtp"
)

const MAX_REDIRECTS = 20

/*
	Namings:
	// converts a string (url) or Dict into an Item
	FetchUnknown: any (a url.URL or Dict) -> Item
		If input is a string, converts it via:
			return FetchURL: url.URL -> Item
		return Construct: Dict -> Item

	// converts user input (webfinger, url, or local file) into Item
	FetchUserInput: string -> Item
		If input starts with @, converts it via:
			ResolveWebfinger: string -> url.URL
		return FetchURL: url.URL -> Item
*/

/*
	Converts a string (url) or Object into an Item
	source represents where the original came from
*/
func FetchUnknown(input any, source *url.URL) (Item, error) {
	switch narrowed := input.(type) {
	case string:
		url, err := url.Parse(narrowed)
		if err != nil {
			return nil, err
		}
		return FetchURL(url)
	case map[string]any:
		return Construct(Object(narrowed), source)
	default:
		return nil, errors.New("can't turn non-string, non-Object into Item")
	}
}

/*
	converts a url into a Object
*/
func FetchURL(link *url.URL) (Item, error) {
	var object Object
	object, err := jtp.Get(
			link,
			`application/activity+json,` +
			`application/ld+json; profile="https://www.w3.org/ns/activitystreams"`,
			[]string{
				"application/activity+json",
				"application/ld+json",
				"application/json",
			},
			MAX_REDIRECTS,
		)

	if err != nil {
		return nil, err
	}

	return Construct(object, link)
}

/*
	converts a Object into an Item
	source is the url whence the Object came
*/
func Construct(object Object, source *url.URL) (Item, error) {
	kind, err := object.GetString("type")
	if err != nil {
		return nil, err
	}

	id, _ := object.GetURL("id")

	if id != nil {
		if source == nil {
			return FetchURL(id)
		}
		if (source.Hostname() != id.Hostname()) || len(object) <= 2 {
			return FetchURL(id)
		}
	}

	switch kind {
	case "Article", "Audio", "Document", "Image", "Note", "Page", "Video":
		return Post{object}, nil

	// case "Create", "Announce", "Dislike", "Like":
	//	return Activity(o), nil

	case "Application", "Group", "Organization", "Person", "Service":
		return Actor{object}, nil

	case "Link":
		return Link{object}, nil

	case "Collection", "OrderedCollection", "CollectionPage", "OrderedCollectionPage":
		return Collection{object, 0}, nil

	default:
		return nil, errors.New("ActivityPub Type " + kind + " is not supported")
	}
}

func FetchUserInput(text string) (Item, error) {
	if strings.HasPrefix(text, "@") {
		link, err := ResolveWebfinger(text)
		if err != nil {
			return nil, err
		}
		return FetchURL(link)
	}

	// if strings.HasPrefix(text, "/") ||
	// 	strings.HasPrefix(text, "./") ||
	// 	strings.HasPrefix(text, "../") {
	// 	file, err := os.Open(text)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	var dictionary Dict
	// 	json.NewEncoder(file).Decode(&dictionary)
	// 	return Construct(dictionary, nil)
	// }

	link, err := url.Parse(text)
	if err != nil {
		return nil, err
	}
	return FetchURL(link)
}

/*
	converts a webfinger identifier to a url
	see: https://datatracker.ietf.org/doc/html/rfc7033
*/
func ResolveWebfinger(username string) (*url.URL, error) {
	username = strings.TrimPrefix(username, "@")
	split := strings.Split(username, "@")
	var account, domain string
	if len(split) != 2 {
		return nil, errors.New("webfinger address must have a separating @ symbol")
	}
	account = split[0]
	domain = split[1]

	query := url.Values{}
	query.Add("resource", "acct:" + account + "@" + domain)
	query.Add("rel", "self")

	link := &url.URL{
		Scheme: "https",
		Host: domain,
		Path: "/.well-known/webfinger",
		RawQuery: query.Encode(),
	}

	response, err := jtp.Get(link, "application/jrd+json", []string{"application/jrd+json"}, MAX_REDIRECTS)
	object := Object(response)

	jrdLinks, err := object.GetList("links")
	if err != nil {
		return nil, err
	}

	var underlyingLink *url.URL = nil

	for _, el := range jrdLinks {
		jrdLink, ok := el.(Object)
		if ok {
			rel, err := jrdLink.GetString("rel")
			if err != nil { continue }
			if rel != "self" { continue }
			mediaType, err := jrdLink.GetMediaType("type")
			if err != nil { continue }
			if !mediaType.Matches([]string{"application/jrd+json", "application/json"}) {
				continue
			}
			href, err := jrdLink.GetURL("href")
			if err != nil { continue }
			underlyingLink = href
			break
		}
	}

	if underlyingLink == nil {
		return nil, errors.New("no matching href was found in the links array of " + link.String())
	}

	return underlyingLink, nil
}
