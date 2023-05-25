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
	"golang.org/x/sync/singleflight"
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
		obj, source, err = FetchURL(url)
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
			obj, source, err = FetchURL(id)
			if err != nil { return nil, nil, err }
		} else if (source.Host != id.Host) || len(obj) <= 2 {
			obj, source, err = FetchURL(id)
			if err != nil { return nil, nil, err }
		}
	}

	// TODO: need to recheck that the id is now accurate, return
	// error if not

	return obj, id, err
}

var group singleflight.Group
type bundle struct {
	item map[string]any
	source *url.URL
	err error
}

/* A map of mutexes is used to ensure no two requests are made simultaneously.
   Instead, the subsequent ones will wait for the first one to finish (and will
   then naturally find its result in the cache) */

func FetchURL(uri *url.URL) (object.Object, *url.URL, error) {
	uriString := uri.String()
	b, _, _ := group.Do(uriString, func() (any, error) {
		json, source, err := 
			jtp.Get(
				uri,
				`application/activity+json,` +
				`application/ld+json; profile="https://www.w3.org/ns/activitystreams"`,
				[]string{
					"application/activity+json",
					"application/ld+json",
					"application/json",
				},
				MAX_REDIRECTS,
			)
		return bundle {
			item: json,
			source: source,
			err: err,
		}, nil
	})
	/* By this point the result has been cached in the LRU cache,
	   so it can be dropped from the singleflight cache */
	group.Forget(uriString)
	return b.(bundle).item, b.(bundle).source, b.(bundle).err
}

/*
	converts a webfinger identifier to a url
	see: https://datatracker.ietf.org/doc/html/rfc7033
*/
func ResolveWebfinger(username string) (string, error) {
	if len(username) == 0 || username[0] != '@' {
		panic("webfinger usernames must begin with @")
	}

	username = username[1:]
	split := strings.SplitN(username, "@", 2)
	var account, domain string
	if len(split) != 2 {
		return "", errors.New("webfinger address must have a separating @ symbol")
	}
	account = split[0]
	domain = split[1]

	link := &url.URL{
		Scheme: "https",
		Host: domain,
		Path: "/.well-known/webfinger",
		RawQuery: (url.Values {
			"resource": []string{"acct:" + account + "@" + domain},
		}).Encode(),
	}

	json, _, err := jtp.Get(link, "application/jrd+json", []string{
		"application/jrd+json",
		"application/json",
	}, MAX_REDIRECTS)
	if err != nil {
		return "", err
	}
	response := object.Object(json)

	jrdLinks, err := response.GetList("links")
	if err != nil {
		return "", err
	}

	found := false
	var underlyingLink string

	for _, el := range jrdLinks {
		asMap, ok := el.(map[string]any)
		o := object.Object(asMap)
		if ok {
			rel, err := o.GetString("rel")
			if err != nil {
				return "", err
			}
			if rel != "self" {
				continue
			}
			mediaType, err := o.GetMediaType("type")
			if errors.Is(err, object.ErrKeyNotPresent) {
				continue
			} else if err != nil {
				return "", err
			}
			if !mediaType.Matches([]string{
				"application/activity+json",
				"application/ld+json",
			}) {
				continue
			}
			href, err := o.GetString("href")
			if err != nil {
				return "", err
			}
			found = true
			underlyingLink = href
			break
		} else {
			return "", fmt.Errorf("unrecognized type %T found in webfinger response", el)
		}
	}

	if !found {
		return "", errors.New("no matching href was found in the links array of " + link.String())
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
