package pub

import (
	"strings"
	"net/url"
	"strconv"
)

type Collection struct {
	Object

	// index *within the current page*
	index int
}

func (c Collection) Kind() string {
	kind, err := c.GetString("type")
	if err != nil { panic(err) }
	return strings.ToLower(kind)
}

func (c Collection) Category() string {
	return "collection"
}

func (c Collection) Identifier() (*url.URL, error) {
	return c.GetURL("id")
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
			// TODO: add a beautiful message here saying
			// failed to load comment: <error>
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
	value, err := c.GetNumber("totalItems")
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(value, 10), nil
}

func (c Collection) items() []any {
	if c.Has("items") {
		if list, err := c.GetList("items"); err == nil {
			return list
		} else {
			return []any{}
		}
	}
	if c.Has("orderedItems") {
		if list, err := c.GetList("orderedItems"); err == nil {
			return list
		} else {
			return []any{}
		}
	}
	return []any{}
}

func (c *Collection) Next() (Item, error) {
	c.index += 1
	return c.Current()
}

func (c *Collection) Previous() (Item, error) {
	c.index -= 1
	return c.Current()
}

/* This return type is a Option<Result<Item>>
   where nil, nil represents None (end of collection)
   nil, err represents Some(Err()) (current item failed construction)
   x, nil represent Some(x) (current item)
   and x, err is invalid */
func (c *Collection) Current() (Item, error) {
	items := c.items()
	if len(items) == 0 {
		kind := c.Kind()
		/* If it is a collection, get the first page */
		if kind == "collection" || kind == "orderedcollection" {
			first, firstErr := c.GetCollection("first")
			if firstErr != nil {
				return nil, nil
			}
			c.Object = first.Object
			c.index = 0
			return c.Current()
		}
	}

	/* This means we are beyond the end of this page */
	if c.index >= len(items) {
		next, err := c.GetCollection("next")
		if err != nil {
			return nil, nil
		}
		c.Object = next.Object
		c.index = 0
		/* Call recursively because the next page may be empty */
		return c.Current()
	} else if c.index < 0 {
		prev, err := c.GetCollection("prev")
		if err != nil {
			return nil, nil
		}
		c.Object = prev.Object
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
