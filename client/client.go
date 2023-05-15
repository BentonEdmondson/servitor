package client

import (
	"errors"
	"net/url"
	"strings"
	"mimicry/jtp"
	"os"
	"encoding/json"
	"mimicry/object"
	"fmt"
)

const MAX_REDIRECTS = 20

func FetchUnknown(input any, source *url.URL) (object.Object, *url.URL, error) {
	var obj object.Object
	switch narrowed := input.(type) {
	case string:
		url, err := url.Parse(narrowed)
		if err != nil {
			return nil, nil, err
		}
		obj, err = FetchURL(url)
		if err != nil { return nil, nil, err }
	case map[string]any:
		obj = object.Object(narrowed)
	default:
		return nil, nil, fmt.Errorf("can't turn non-string, non-object %T into Item", input)
	}

	id, err := obj.GetURL("id")
	if errors.Is(err, object.ErrKeyNotPresent) {
		id = nil
		err = nil
	} else if err != nil {
		return nil, nil, err
	}

	if id != nil {
		if source == nil {
			obj, err = FetchURL(id)
			if err != nil { return nil, nil, err }
		} else if (source.Host != id.Host) || len(obj) <= 2 {
			obj, err = FetchURL(id)
			if err != nil { return nil, nil, err }
		}
	}

	// TODO: need to recheck that the id is now accurate, return
	// error if not

	return obj, id, err
}

func FetchURL(link *url.URL) (object.Object, error) {
	return jtp.Get(
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

	json, err := jtp.Get(link, "application/jrd+json", []string{"application/jrd+json"}, MAX_REDIRECTS)
	response := object.Object(json)

	jrdLinks, err := response.GetList("links")
	if err != nil {
		return nil, err
	}

	var underlyingLink *url.URL = nil

	for _, el := range jrdLinks {
		jrdLink, ok := el.(object.Object)
		if ok {
			rel, err := jrdLink.GetString("rel")
			if err != nil { continue }
			if rel != "self" { continue }
			mediaType, err := jrdLink.GetMediaType("type")
			if err != nil { continue }
			if !mediaType.Matches([]string{"application/activity+json"}) {
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

func FetchFromFile(name string) (object.Object, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	var obj object.Object
	json.NewDecoder(file).Decode(&obj)
	return obj, nil
}
