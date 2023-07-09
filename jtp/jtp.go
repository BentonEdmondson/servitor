package jtp

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru/v2"
	"servitor/mime"
	"net"
	"net/url"
	"regexp"
	"strings"
	"servitor/config"
)

var dialer = &net.Dialer{
	Timeout: config.Parsed.Network.Timeout,
}

type bundle struct {
	item   map[string]any
	source *url.URL
	err    error
}

var cache, _ = lru.New[string, bundle](config.Parsed.Network.CacheSize)

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
func Get(link *url.URL, accept string, tolerated []string, maxRedirects uint) (map[string]any, *url.URL, error) {
	if cached, ok := cache.Get(link.String()); ok {
		return cached.item, cached.source, cached.err
	}

	if link.Scheme != "https" {
		return nil, nil, errors.New(link.Scheme + " is not supported in requests, only https")
	}

	port := link.Port()
	if port == "" {
		port = "443"
	}

	hostport := net.JoinHostPort(link.Hostname(), port)

	connection, err := tls.DialWithDialer(dialer, "tcp", hostport, nil)
	if err != nil {
		return nil, nil, err
	}

	_, err = connection.Write([]byte(
		"GET " + link.RequestURI() + " HTTP/1.0\r\n" +
			"Host: " + link.Host + "\r\n" +
			"Accept: " + accept + "\r\n" +
			"\r\n",
	))
	if err != nil {
		return nil, nil, errors.Join(err, connection.Close())
	}

	buf := bufio.NewReader(connection)
	statusLine, err := buf.ReadString('\n')
	if err != nil {
		return nil, nil, errors.Join(
			fmt.Errorf("failed to parse HTTP status line: %w", err),
			connection.Close(),
		)
	}

	status, err := parseStatusLine(statusLine)
	if err != nil {
		return nil, nil, errors.Join(err, connection.Close())
	}

	if strings.HasPrefix(status, "3") {
		location, err := findLocation(buf, link)
		if err != nil {
			return nil, nil, errors.Join(err, connection.Close())
		}

		if maxRedirects == 0 {
			return nil, nil, errors.Join(
				errors.New("received "+status+" after redirecting too many times"),
				connection.Close(),
			)
		}

		if err := connection.Close(); err != nil {
			return nil, nil, err
		}
		var b bundle
		b.item, b.source, b.err = Get(location, accept, tolerated, maxRedirects-1)
		cache.Add(link.String(), b)
		return b.item, b.source, b.err
	}

	if status != "200" && status != "201" && status != "202" && status != "203" {
		return nil, nil, errors.Join(
			errors.New("received invalid status "+status),
			connection.Close(),
		)
	}

	err = validateHeaders(buf, tolerated)
	if err != nil {
		return nil, nil, errors.Join(err, connection.Close())
	}

	var dictionary map[string]any
	err = json.NewDecoder(buf).Decode(&dictionary)
	if err != nil {
		return nil, nil, errors.Join(
			fmt.Errorf("failed to parse JSON: %w", err),
			connection.Close(),
		)
	}

	if err := connection.Close(); err != nil {
		return nil, nil, err
	}

	cache.Add(link.String(), bundle{
		item:   dictionary,
		source: link,
		err:    nil,
	})
	return dictionary, link, nil
}

func parseStatusLine(text string) (string, error) {
	matches := statusLineRegexp.FindStringSubmatch(text)

	if len(matches) != 2 {
		return "", errors.New("received invalid status line: " + text)
	}

	return matches[1], nil
}

func parseContentType(text string) (*mime.MediaType, bool, error) {
	matches := contentTypeRegexp.FindStringSubmatch(text)

	if len(matches) != 2 {
		return nil, false, nil
	}

	mediaType, err := mime.Parse(matches[1])
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
			return errors.New("response is of invalid type " + mediaType.Essence)
		}
	}

	if !contentTypeValidated {
		return errors.New("response is missing a content type")
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
	return nil, errors.New("response is missing Location header")
}
