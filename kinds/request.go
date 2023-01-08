package kinds

import (
	"strings"
	"net/http"
	"net/url"
	"errors"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

var client = &http.Client{}
//var cache = TODO

const requiredContentType = `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`
const optionalContentType = "application/activity+json"

func Fetch(url *url.URL) (Content, error) {
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

	if resp.StatusCode != 200 {
		return nil, errors.New("The server returned a status code of " + resp.Status)
	}

	// TODO: delete the pointless first if right here
	// TODO: for the sake of static servers, accept application/json
	//		 as well iff it contains the @context key (but double check
	//		 that it is absolutely necessary)
	if contentType := resp.Header.Get("Content-Type"); contentType == "" {
		return nil, errors.New("The server's response did not contain a content type")
	} else if !strings.Contains(contentType, requiredContentType) && !strings.Contains(contentType, optionalContentType) {
		return nil, errors.New("The server responded with the invalid content type of " + contentType)
	}

	var unstructured map[string]any
	if err := json.Unmarshal(body, &unstructured); err != nil {
		return nil, err
	}

	return Construct(unstructured, url)
}

func FetchWebFinger(username string) (Actor, error) {
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

	content, err := Fetch(underlyingLink)
	if err != nil { return nil, err }

	actor, ok := content.(Actor)
	if !ok { return nil, errors.New("content returned by the WebFinger request was not an Actor") }

	return actor, nil
}

func FetchUnknown(unknown string) (Content, error) {
	if strings.HasPrefix(unknown, "@") {
		return FetchWebFinger(unknown)
	}

	url, err := url.Parse(unknown)
	if err != nil {
		return nil, err
	}

	return Fetch(url)
}
