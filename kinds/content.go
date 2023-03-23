package kinds

import (
	"net/url"
)

// TODO: rename to Item
// TODO: a collection should probably not be an item
type Content interface {
	String(width int) (string, error)
	Preview() (string, error)
	Kind() (string, error)
	Category() string

	// if the id field is absent or nil, then
	// this should return (nil, nil),
	// if it is present and malformed, then use
	// an error
	Identifier() (*url.URL, error)
	Raw() Dict
}