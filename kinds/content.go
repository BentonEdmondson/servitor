package kinds

import (
	"net/url"
)

type Content interface {
	String() (string, error)
	Kind() (string, error)
	Category() string

	// if the id field is absent or nil, then
	// this should return (nil, nil),
	// if it is present and malformed, then use
	// an error
	Identifier() (*url.URL, error)
}