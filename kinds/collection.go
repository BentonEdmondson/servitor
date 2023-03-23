package kinds

import (
	"strings"
	"net/url"
	"strconv"
)

type Collection struct {
	page Dict

	// index *within the current page*
	index int
}

func (c Collection) Raw() Dict {
	return c.page
}

func (c Collection) Kind() (string, error) {
	kind, err := Get[string](c.page, "type")
	return strings.ToLower(kind), err
}

func (c Collection) Category() string {
	return "collection"
}

func (c Collection) Identifier() (*url.URL, error) {
	return GetURL(c.page, "id")
}

func (c Collection) String(width int) (string, error) {
	elements := []string{}

	const elementsToShow = 3
	for len(elements) < elementsToShow {

		current, err := c.Current()
		if current == nil && err == nil {
			break
		}

		if err != nil {
			c.Next()
			continue
		}

		output, err := current.Preview()
		if err != nil {
			return "", err
		}

		elements = append(elements, output)

		c.Next()
	}
	
	return strings.Join(elements, "\n"), nil
}

func (c Collection) Size() (string, error) {
	value, err := Get[float64](c.page, "totalItems")
	if err != nil {
		return "", err
	}
	return strconv.Itoa(int(value)), nil
}

func (c Collection) items() []any {
	itemsList, itemsErr := Get[[]any](c.page, "items")
	if itemsErr == nil {
		return itemsList
	}
	orderedItemsList, orderedItemsErr := Get[[]any](c.page, "orderedItems")
	if orderedItemsErr == nil {
		return orderedItemsList
	}

	return []any{}
}

func (c *Collection) Next() (Content, error) {
	c.index += 1
	return c.Current()
}

func (c *Collection) Previous() (Content, error) {
	c.index -= 1
	return c.Current()
}

/* This return type is a Option<Result<Content>>
   where nil, nil represents None (end of collection)
   nil, err represents Some(Err()) (current item failed construction)
   x, nil represent Some(x) (current item)
   and x, err is invalid */
func (c *Collection) Current() (Content, error) {
	items := c.items()
	if len(items) == 0 {
		kind, kindErr := c.Kind()
		if kindErr != nil {
			return nil, nil
		}

		/* If it is a collection, get the first page */
		if kind == "collection" || kind == "orderedcollection" {
			first, firstErr := GetItem[Collection](c.page, "first")
			if firstErr != nil {
				return nil, nil
			}
			c.page = first.page
			c.index = 0
			return c.Current()
		}
	}

	/* At this point we know items are present */

	/* This means we are beyond the end of this page */
	if c.index >= len(items) {
		next, err := GetItem[Collection](c.page, "next")
		if err != nil {
			return nil, nil
		}
		c.page = next.page
		c.index = 0
		/* Call recursively because the next page may be empty */
		return c.Current()
	} else if c.index < 0 {
		prev, err := GetItem[Collection](c.page, "prev")
		if err != nil {
			return nil, nil
		}
		c.page = prev.page
		items := c.items()
		/* If this new page is empty, this will be -1, and the
		   call to Current will flip back yet another page, as
		   intended */
		c.index = len(items)-1
		return c.Current()
	}

	/* At this point we know index is within items */ 

	id, _ := c.Identifier()

	return FetchUnknown(items[c.index], id)
}

func (c Collection) Preview() (string, error) {
	return "I will get rid of this function", nil
}
