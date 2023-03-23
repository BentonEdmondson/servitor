package jtp

import (
	"regexp"
	"golang.org/x/exp/slices"
	"errors"
	"crypto/tls"
	"net"
	"net/url"
	"bufio"
	"fmt"
	"strings"
	"encoding/json"
)

// TODO: parseMediaType should probably return an error if the mediaType is invalid
// or at least do something that will be easier to debug

type MediaType struct {
	Supertype string
	Subtype string
	/* Full omits the parameters */
	Full string
}

var dialer = &tls.Dialer{
	NetDialer: &net.Dialer{},
}

var mediaTypeRegexp = regexp.MustCompile(`(?s)^(([!#$%&'*+\-.^_\x60|~a-zA-Z0-9]+)/([!#$%&'*+\-.^_\x60|~a-zA-Z0-9]+)).*$`)
var statusLineRegexp = regexp.MustCompile(`^HTTP/1\.[0-9] ([0-9]{3}).*\n$`)
var contentTypeRegexp = regexp.MustCompile(`^(?i:content-type):[ \t\r]*(.*?)[ \t\r]*\n$`)
var locationRegexp = regexp.MustCompile(`^(?i:location):[ \t\r]*(.*?)[ \t\r]*\n$`)

var acceptHeader = `application/activity+json,` +
	`application/ld+json; profile="https://www.w3.org/ns/activitystreams"`
	
var toleratedTypes = []string{
	"application/activity+json",
	"application/ld+json",
	"application/json",
}

/*
	I send an HTTP/1.0 request to ensure the server doesn't respond
	with chunked transfer encoding.
	See: https://httpwg.org/specs/rfc9110.html
*/

/*
	link
		the url being requested
	maxRedirects
		the maximum number of redirects to take
*/
// TODO: the number of redirects must be limited
func Get(link *url.URL, maxRedirects uint) (map[string]any, error) {

	if link.Scheme != "https" {
		return nil, errors.New(link.Scheme + " is not supported in requests, only https")
	}

	port := link.Port()
	if port == "" {
		port = "443"
	}

	hostport := net.JoinHostPort(link.Hostname(), port)

	connection, err := dialer.Dial("tcp", hostport)
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	_, err = connection.Write([]byte(
		"GET " + link.RequestURI() + " HTTP/1.0\r\n" +
		"Host: " + link.Host + "\r\n" +
		"Accept: " + acceptHeader + "\r\n" +
		"Accept-Encoding: identity\r\n" +
		"\r\n",
	))
	if err != nil {
		return nil, err
	}

	buf := bufio.NewReader(connection)
	statusLine, err := buf.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("Encountered error while reading status line of HTTP response: %w", err)
	}

	status, err := parseStatusLine(statusLine)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(status, "3") {
		location, err := findLocation(buf, link)
		if err != nil {
			return nil, err
		}

		if maxRedirects == 0 {
			return nil, errors.New("Received " + status + " but max redirects has already been reached")
		}

		return Get(location, maxRedirects - 1)
	}

	if status != "200" && status != "201" && status != "202" && status != "203" {
		return nil, errors.New("Received invalid status " + status)
	}

	err = validateHeaders(buf)
	if err != nil {
		return nil, err
	}

	var dictionary map[string]any
	err = json.NewDecoder(buf).Decode(&dictionary)
	if err != nil {
		return nil, err
	}

	return dictionary, nil
}

func ParseMediaType(text string) (MediaType, error) {
	matches := mediaTypeRegexp.FindStringSubmatch(text)

	if len(matches) != 4 {
		return MediaType{}, errors.New(text + " is not a valid media type")
	}

	return MediaType{
		Supertype: matches[2],
		Subtype: matches[3],
		Full: matches[1],
	}, nil
}

func parseStatusLine(text string) (string, error) {
	matches := statusLineRegexp.FindStringSubmatch(text)

	if len(matches) != 2 {
		return "", errors.New("Received invalid status line: " + text)
	}

	return matches[1], nil
}

func parseContentType(text string) (MediaType, bool, error) {
	matches := contentTypeRegexp.FindStringSubmatch(text)

	if len(matches) != 2 {
		return MediaType{}, false, nil
	}

	mediaType, err := ParseMediaType(matches[1])
	if err != nil {
		return MediaType{}, true, err
	}

	return mediaType, true, nil
}

func parseLocation(text string, baseLink *url.URL) (link *url.URL, isLocationLine bool, err error) {
	matches := locationRegexp.FindStringSubmatch(text)

	if len(matches) != 2 {
		return nil, false, nil
	}

	reference, err := url.Parse(matches[1])
	if err != nil {
		return nil, true, err
	}

	return baseLink.ResolveReference(reference), true, nil
}

func validateHeaders(buf *bufio.Reader) error {
	contentTypeValidated := false
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			return err
		}

		if line == "\r\n" || line == "\n" {
			break
		}

		mediaType, isContentTypeLine, err := parseContentType(line)
		if err != nil {
			return err
		}
		if !isContentTypeLine {
			continue
		}

		if slices.Contains(toleratedTypes, mediaType.Full) {
			contentTypeValidated = true
		} else {
			return errors.New("Response contains invalid content type " + mediaType.Full)
		}
	}

	if !contentTypeValidated {
		return errors.New("Response did not contain a content type")
	}

	return nil
}

func findLocation(buf *bufio.Reader, baseLink *url.URL) (*url.URL, error) {
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			return nil, err
		}

		if line == "\r\n" || line == "\n" {
			break
		}

		location, isLocationLine, err := parseLocation(line, baseLink)
		if err != nil {
			return nil, err
		}
		if !isLocationLine {
			continue
		}

		return location, nil
	}
	return nil, errors.New("Location is not present in headers")
}
