package request

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

func Fetch(link *url.URL) (map[string]any, error) {
	const requiredContentType = `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`
	const optionalContentType = "application/activity+json"

	// convert URL to string
	url := link.String()

	// create the get request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return map[string]any{}, err
	}

	// add the accept header
	// § 3.2
	req.Header.Add("Accept", fmt.Sprintf("%s, %s", requiredContentType, optionalContentType))

	// send the request
	resp, err := client.Do(req)
	if err != nil {
		return map[string]any{}, err
	}

	// check the status code
	if resp.StatusCode != 200 {
		return nil, errors.New("The server returned a status code of " + string(resp.StatusCode))
	}

	// check the response content type
	if contentType := resp.Header.Get("Content-Type"); contentType == "" {
		return nil, errors.New("The server's response did not contain a content type")
	} else if !strings.Contains(contentType, requiredContentType) && !strings.Contains(contentType, optionalContentType) {
		return nil, errors.New("The server responded with the invalid content type of " + contentType)
	}

	// read the body into a map
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var object map[string]any
	if err := json.Unmarshal(body, &object); err != nil {
		return nil, err
	}

	return object, nil
}
