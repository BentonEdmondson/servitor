package pub

import (
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"servitor/client"
	"servitor/object"
	"net/url"
	"sync"
)

/*
	Methods are:
	Category
	Kind
	Identifier
	Next
	Size
	Items (returns list)
	String // maybe just show this page, and Next can be a button
		// the infiniscroll will be provided by the View package
*/

// Should probably take in a constructor, actor gives NewActivity
// and Post gives NewPost, but not exactly, they can wrap them
// in a function which also checks whether the element is
// valid in the given context

type Collection struct {
	kind string
	id   *url.URL

	elements    []any
	elementsErr error
	next        any
	nextErr     error

	size    uint64
	sizeErr error
}

func NewCollection(input any, source *url.URL) (*Collection, error) {
	o, id, err := client.FetchUnknown(input, source)
	if err != nil {
		return nil, err
	}
	return NewCollectionFromObject(o, id)
}

func NewCollectionFromObject(o object.Object, id *url.URL) (*Collection, error) {
	c := &Collection{}
	c.id = id
	var err error
	if c.kind, err = o.GetString("type"); err != nil {
		return nil, err
	}

	if !slices.Contains([]string{
		"Collection", "OrderedCollection", "CollectionPage", "OrderedCollectionPage",
	}, c.kind) {
		return nil, fmt.Errorf("%w: %s is not a Collection", ErrWrongType, c.kind)
	}

	if c.kind == "Collection" || c.kind == "CollectionPage" {
		c.elements, c.elementsErr = o.GetList("items")
	} else {
		c.elements, c.elementsErr = o.GetList("orderedItems")
	}

	if c.kind == "Collection" || c.kind == "OrderedCollection" {
		c.next, c.nextErr = o.GetAny("first")
	} else {
		c.next, c.nextErr = o.GetAny("next")
	}

	c.size, c.sizeErr = o.GetNumber("totalItems")

	return c, nil
}

func (c *Collection) Size() (uint64, error) {
	return c.size, c.sizeErr
}

func (c *Collection) Harvest(amount uint, startingPoint uint) ([]Tangible, Container, uint) {
	return c.harvestWithEmptyCount(amount, startingPoint, 0)
}

func (c *Collection) harvestWithEmptyCount(amount uint, startingPoint uint, emptyCount int) ([]Tangible, Container, uint) {
	if c == nil {
		panic("can't harvest nil collection")
	}

	if c.elementsErr != nil && !errors.Is(c.elementsErr, object.ErrKeyNotPresent) {
		return []Tangible{NewFailure(c.elementsErr)}, nil, 0
	}

	var length uint
	if errors.Is(c.elementsErr, object.ErrKeyNotPresent) {
		length = 0
	} else {
		length = uint(len(c.elements))
	}

	if length == 0 {
		emptyCount += 1
	}

	/*
		This is set at 3 because 3 seems to be the maximum amount that servers send, besides cases of infinite loops.
		Mastodon sends 3 in the following case:
			- the first is the Collection itself, which has no items because the items are in CollectionPages
			- the next page (the first CollectionPage) is empty because it only holds self-replies and there are none
			- the next page (the second CollectionPage) is empty because it holds replies from others and there are none
	*/
	if emptyCount > 3 {
		return []Tangible{NewFailure(errors.New("refusing to read the next collection because >3 consecutive empty collections have been encountered"))}, nil, 0
	}

	var amountFromThisPage uint
	if startingPoint >= length {
		amountFromThisPage = 0
	} else if length > amount+startingPoint {
		amountFromThisPage = amount
	} else {
		amountFromThisPage = length - startingPoint
	}

	fromThisPage := make([]Tangible, amountFromThisPage)
	var fromLaterPages []Tangible
	var nextCollection Container
	var nextStartingPoint uint

	var wg sync.WaitGroup
	for i := uint(0); i < amountFromThisPage; i++ {
		i := i
		wg.Add(1)
		go func() {
			fromThisPage[i] = NewTangible(c.elements[i+startingPoint], c.id)
			wg.Done()
		}()
	}

	wg.Add(1)
	go func() {
		if length > amount+startingPoint {
			fromLaterPages, nextCollection, nextStartingPoint = []Tangible{}, c, amount+startingPoint
		} else if errors.Is(c.nextErr, object.ErrKeyNotPresent) {
			fromLaterPages, nextCollection, nextStartingPoint = []Tangible{}, nil, 0
		} else if c.nextErr != nil {
			fromLaterPages, nextCollection, nextStartingPoint = []Tangible{NewFailure(c.nextErr)}, nil, 0
		} else if next, err := NewCollection(c.next, c.id); err != nil {
			fromLaterPages, nextCollection, nextStartingPoint = []Tangible{NewFailure(err)}, nil, 0
		} else {
			fromLaterPages, nextCollection, nextStartingPoint = next.harvestWithEmptyCount(amount-amountFromThisPage, 0, emptyCount)
		}

		wg.Done()
	}()

	wg.Wait()

	return append(fromThisPage, fromLaterPages...), nextCollection, nextStartingPoint
}
