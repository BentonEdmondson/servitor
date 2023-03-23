package kinds

import (
	"errors"
	"net/url"
	"strings"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

/*
	TODO: updated plan:
	I need a function which accepts a string (url) or dict and converts
	it into an Item (Currently under GetContent)

	I need another function which accepts a string (webfinger or url) and converts
	it into an Item (currently under FetchUnkown)

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

var client = &http.Client{}

const requiredContentType = `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`
const optionalContentType = "application/activity+json"

func FetchUnknown(input any, source *url.URL) (Content, error) {
	switch narrowed := input.(type) {
	case string:
		// TODO: detect the 3 `Public` identifiers and error on them
		url, err := url.Parse(narrowed)
		if err != nil {
			return nil, err
		}
		return FetchURL(url)
	case Dict:
		return Construct(narrowed, source)
	default:
		return nil, errors.New("Can't resolve non-string, non-Dict into Item.")
	}
}

func FetchURL(url *url.URL) (Content, error) {
	link := url.String()

	req, err := http.NewRequest("GET", link, nil) // `nil` is body
	if err != nil {
		return nil, err
	}

	// add the accept header, some servers only respond if the optional
	// content type is included as well
	// § 3.2
	req.Header.Add("Accept", fmt.Sprintf("%s, %s", requiredContentType, optionalContentType))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// GNU Social servers return 202 (with a correct body) instead of 200
	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		return nil, errors.New("The server returned a status code of " + resp.Status)
	}

	// TODO: delete the pointless first if right here
	// TODO: for the sake of static servers, accept application/json
	//		 as well iff it contains the @context key (but double check
	//		 that it is absolutely necessary)
	if contentType := resp.Header.Get("Content-Type"); contentType == "" {
		return nil, errors.New("The server's response did not contain a content type")

	// TODO: accept application/ld+json, application/json, and application/activity+json as responses
	} else if !strings.Contains(contentType, requiredContentType) && !strings.Contains(contentType, optionalContentType) {
		return nil, errors.New("The server responded with the invalid content type of " + contentType)
	}

	var unstructured map[string]any
	if err := json.Unmarshal(body, &unstructured); err != nil {
		return nil, err
	}

	return Construct(unstructured, url)
}

// TODO, add a verbose debugging output mode
// to debug problems that arise with this thing
// looping too much and whatnot

// `unstructured` is the JSON to construct from,
// source is where the JSON was received from,
// used to ensure the reponse is trustworthy
func Construct(unstructured Dict, source *url.URL) (Content, error) {
	kind, err := Get[string](unstructured, "type")
	if err != nil {
		return nil, err
	}

	// this requirement should be removed, and the below check
	// should be checking if only type or only type and id
	// are present on the element
	hasIdentifier := true
	id, err := GetURL(unstructured, "id")
	if err != nil {
		hasIdentifier = false
	}

	// if the JSON came from a source (e.g. inline in another collection), with a
	// different hostname than its ID, refetch
	// if the JSON only has two keys (type and id), refetch
	if (source != nil && id != nil) {
		if (source.Hostname() != id.Hostname()) || (len(unstructured) <= 2 && hasIdentifier) {
			return FetchURL(id)
		}
	}

	switch kind {
	case "Article", "Audio", "Document", "Image", "Note", "Page", "Video":
		// TODO: figure out the way to do this directly
		post := Post{}
		post = unstructured
		return post, nil

	// case "Create":
	// 	fallthrough
	// case "Announce":
	// 	fallthrough
	// case "Dislike":
	// 	fallthrough
	// case "Like":
	// 	fallthrough
	// case "Question":
	// 	return Activity{unstructured}, nil

	case "Application", "Group", "Organization", "Person", "Service":
		// TODO: nicer way to do this?
		actor := Actor{}
		actor = unstructured
		return actor, nil

	case "Link":
		link := Link{}
		link = unstructured
		return link, nil

	case "Collection", "OrderedCollection", "CollectionPage", "OrderedCollectionPage":
		collection := Collection{}
		collection = Collection{unstructured, 0}
		return collection, nil

	default:
		return nil, errors.New("Object of Type " + kind + " unsupported")
	}
}

func FetchUserInput(text string) (Content, error) {
	if strings.HasPrefix(text, "@") {
		link, err := ResolveWebfinger(text)
		if err != nil {
			return nil, err
		}
		return FetchURL(link)
	} else {
		link, err := url.Parse(text)
		if err != nil {
			return nil, err
		}
		return FetchURL(link)
	}
}

func ResolveWebfinger(username string) (*url.URL, error) {
	// description of WebFinger: https://www.rfc-editor.org/rfc/rfc7033.html

	username = strings.TrimPrefix(username, "@")

	split := strings.Split(username, "@")
	var account, domain string
	if len(split) != 2 {
		return nil, errors.New("webfinger address must have a separating @ symbol")
	} else {
		account = split[0]
		domain = split[1]
	}

	query := url.Values{}
	query.Add("resource", fmt.Sprintf("acct:%s@%s", account, domain))
	query.Add("rel", "self")

	link := url.URL{
		Scheme: "https",
		Host: domain,
		Path: "/.well-known/webfinger",
		RawQuery: query.Encode(),
	}

	req, err := http.NewRequest("GET", link.String(), nil) // `nil` is body
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("the server responded to the WebFinger query %s with %s", link.String(), resp.Status))
	} else if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "application/jrd+json") && !strings.Contains(contentType, "application/json") {
		return nil, errors.New("the server responded to the WebFinger query with invalid Content-Type " + contentType)
	}

	var jrd Dict
	if err := json.Unmarshal(body, &jrd); err != nil {
		return nil, err
	}

	jrdLinks, err := GetList(jrd, "links")
	if err != nil {
		return nil, err
	}

	var underlyingLink *url.URL = nil

	for _, el := range jrdLinks {
		jrdLink, ok := el.(Dict)
		if ok {
			rel, err := Get[string](jrdLink, "rel")
			if err != nil { continue }
			if rel != "self" { continue }
			mediaType, err := Get[string](jrdLink, "type")
			if err != nil { continue }
			if !strings.Contains(mediaType, requiredContentType) && !strings.Contains(mediaType, optionalContentType) {
				continue
			}
			href, err := GetURL(jrdLink, "href")
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
