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
var contentTypeRegexp = regexp.MustCompile(`^(?i:content-type:)[ \t\r]*(.*?)[ \t\r]*\n$`)

/*
	I send an HTTP/1.0 request to ensure the server doesn't respond
	with chunked transfer encoding.
	See: https://httpwg.org/specs/rfc9110.html
*/

/*
	link
		the url being requested
	requestedTypes
		the `Accept` header value
	toleratedTypes
		a list of media types (excluding parameters) that
		should be accepted in the response, all other media
		types result in an error
*/
// TODO: the number of redirects must be limited
func Get(link *url.URL, requestedTypes string, toleratedTypes []string) (map[string]any, error) {

	if link.Scheme != "https" {
		return nil, errors.New(link.Scheme + "is not supported in requests, only https")
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
		"Accept: " + requestedTypes + "\r\n" +
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
		return nil, errors.New("Return code " + status + ", I haven't implemented redirects yet")
	}

	if status != "200" && status != "201" && status != "202" && status != "203" {
		return nil, errors.New("Received invalid status " + status)
	}

	err = validateHeaders(buf, toleratedTypes)
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

func ParseMediaType(text string) MediaType {
	matches := mediaTypeRegexp.FindStringSubmatch(text)

	if len(matches) != 4 {
		return MediaType{}
	}

	return MediaType{
		Supertype: matches[2],
		Subtype: matches[3],
		Full: matches[1],
	}
}

func parseStatusLine(text string) (string, error) {
	matches := statusLineRegexp.FindStringSubmatch(text)

	if len(matches) != 2 {
		return "", errors.New("Received invalid status line: " + text)
	}

	return matches[1], nil
}

func parseContentType(text string) (mediaType MediaType, isContentTypeLine bool) {
	matches := contentTypeRegexp.FindStringSubmatch(text)

	if len(matches) != 2 {
		return MediaType{}, false
	}

	return ParseMediaType(matches[1]), true
}

func validateHeaders(buf *bufio.Reader, toleratedTypes []string) error {
	contentTypeValidated := false
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			return err
		}

		if line == "\r\n" {
			break
		}

		mediaType, isContentTypeLine := parseContentType(line)
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
