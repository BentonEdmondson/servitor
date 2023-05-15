package jtp

import (
	"regexp"
	"errors"
	"crypto/tls"
	"net"
	"net/url"
	"bufio"
	"fmt"
	"strings"
	"encoding/json"
)

var dialer = &tls.Dialer{
	NetDialer: &net.Dialer{},
}

var mediaTypeRegexp = regexp.MustCompile(`(?s)^(([!#$%&'*+\-.^_\x60|~a-zA-Z0-9]+)/([!#$%&'*+\-.^_\x60|~a-zA-Z0-9]+)).*$`)
var statusLineRegexp = regexp.MustCompile(`^HTTP/1\.[0-9] ([0-9]{3}).*\n$`)
var contentTypeRegexp = regexp.MustCompile(`^(?i:content-type):[ \t\r]*(.*?)[ \t\r]*\n$`)
var locationRegexp = regexp.MustCompile(`^(?i:location):[ \t\r]*(.*?)[ \t\r]*\n$`)

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
func Get(link *url.URL, accept string, tolerated []string, maxRedirects uint) (map[string]any, error) {
	if link.Scheme != "https" {
		return nil, errors.New(link.Scheme + " is not supported in requests, only https")
	}

	port := link.Port()
	if port == "" {
		port = "443"
	}

	// TODO: link.Host may work instead of needing net.JoinHostPort
	hostport := net.JoinHostPort(link.Hostname(), port)

	connection, err := dialer.Dial("tcp", hostport)
	if err != nil {
		return nil, err
	}

	_, err = connection.Write([]byte(
		"GET " + link.RequestURI() + " HTTP/1.0\r\n" +
		"Host: " + link.Host + "\r\n" +
		"Accept: " + accept + "\r\n" +
		"\r\n",
	))
	if err != nil {
		return nil, errors.Join(err, connection.Close())
	}

	buf := bufio.NewReader(connection)
	statusLine, err := buf.ReadString('\n')
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to parse HTTP status line: %w", err),
			connection.Close(),
		)
	}

	status, err := parseStatusLine(statusLine)
	if err != nil {
		return nil, errors.Join(err, connection.Close())
	}

	if strings.HasPrefix(status, "3") {
		location, err := findLocation(buf, link)
		if err != nil {
			return nil, errors.Join(err, connection.Close())
		}

		if maxRedirects == 0 {
			return nil, errors.Join(
				errors.New("Received " + status + " but max redirects has already been reached"),
				connection.Close(),
			)
		}

		if err := connection.Close(); err != nil {
			return nil, err
		}
		return Get(location, accept, tolerated, maxRedirects - 1)
	}

	if status != "200" && status != "201" && status != "202" && status != "203" {
		return nil, errors.Join(
			errors.New("received invalid status " + status),
			connection.Close(),
		)
	}

	err = validateHeaders(buf, tolerated)
	if err != nil {
		return nil, errors.Join(err, connection.Close())
	}

	var dictionary map[string]any
	err = json.NewDecoder(buf).Decode(&dictionary)
	if err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to parse JSON: %w", err),
			connection.Close(),
		)
	}

	if err := connection.Close(); err != nil {
		return nil, err
	}

	return dictionary, nil
}

func parseStatusLine(text string) (string, error) {
	matches := statusLineRegexp.FindStringSubmatch(text)

	if len(matches) != 2 {
		return "", errors.New("Received invalid status line: " + text)
	}

	return matches[1], nil
}

func parseContentType(text string) (*MediaType, bool, error) {
	matches := contentTypeRegexp.FindStringSubmatch(text)

	if len(matches) != 2 {
		return nil, false, nil
	}

	mediaType, err := ParseMediaType(matches[1])
	if err != nil {
		return nil, true, err
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

func validateHeaders(buf *bufio.Reader, tolerated []string) error {
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

		if mediaType.Matches(tolerated) {
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

type MediaType struct {
	Supertype string
	Subtype string
	/* Full omits the parameters */
	Full string
}

func ParseMediaType(text string) (*MediaType, error) {
	matches := mediaTypeRegexp.FindStringSubmatch(text)

	if len(matches) != 4 {
		return nil, errors.New(text + " is not a valid media type")
	}

	return &MediaType{
		Supertype: matches[2],
		Subtype: matches[3],
		Full: matches[1],
	}, nil
}

func (m *MediaType) Matches(mediaTypes []string) bool {
	for _, mediaType := range mediaTypes {
		if m.Full == mediaType {
			return true
		}
	}
	return false
}
