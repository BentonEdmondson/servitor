package kinds

import (
	"net/url"
)

type Link Dict

func (l Link) Kind() (string, error) {
	return "link", nil
}

func (l Link) Category() string {
	return "link"
}

func (l Link) MediaType() (string, error) {
	return Get[string](l, "mediaType")
}

func (l Link) URL() (*url.URL, error) {
	return GetURL(l, "href")
}

func (l Link) Alt() (string, error) {
	return Get[string](l, "name")
}