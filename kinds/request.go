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

func Fetch(url *url.URL) (Content, error) {
	const requiredContentType = `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`
	const optionalContentType = "application/activity+json"

	link := url.String()

	req, err := http.NewRequest("GET", link, nil)
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

	if resp.StatusCode != 200 {
		return nil, errors.New("The server returned a status code of " + string(resp.StatusCode))
	}

	if contentType := resp.Header.Get("Content-Type"); contentType == "" {
		return nil, errors.New("The server's response did not contain a content type")
	} else if !strings.Contains(contentType, requiredContentType) && !strings.Contains(contentType, optionalContentType) {
		return nil, errors.New("The server responded with the invalid content type of " + contentType)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var unstructured map[string]any
	if err := json.Unmarshal(body, &unstructured); err != nil {
		return nil, err
	}

	return Construct(unstructured, url)
}
