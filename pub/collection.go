package pub

import (
	"net/url"
	"mimicry/object"
	"errors"
	"mimicry/client"
	"fmt"
	"golang.org/x/exp/slices"
	"log"
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
	id *url.URL

	elements []any; elementsErr error
	next any; nextErr error

	size uint64; sizeErr error
}

func NewCollection(input any, source *url.URL) (*Collection, error) {
	c := &Collection{}
	var o object.Object; var err error
	o, c.id, err = client.FetchUnknown(input, source)
	if err != nil { return nil, err }
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

func (c *Collection) Kind() string {
	return c.kind
}

func (c *Collection) Size() (uint64, error) {
	return c.size, c.sizeErr
}

func (c *Collection) Harvest(amount uint, startingPoint uint) ([]Tangible, Container, uint) {
	// To work through this problem you need to go through this step by step and
	// make sure the logic is good. Then you should probably start writing some tests
	
	log.Printf("amount: %d starting: %d", amount, startingPoint)
	if c.elementsErr != nil && !errors.Is(c.elementsErr, object.ErrKeyNotPresent) {
		return []Tangible{NewFailure(c.elementsErr)}, nil, 0
	}

	var length uint
	if errors.Is(c.elementsErr, object.ErrKeyNotPresent) {
		length = 0
	} else {
		length = uint(len(c.elements))
	}
	log.Printf("length: %d", length)

	// TODO: change to bool nextWillBeFetched in which case amount from this page is all
	// and later on the variable is clear

	var amountFromThisPage uint
	if startingPoint >= length {
		amountFromThisPage = 0
	} else if length > amount + startingPoint {
		amountFromThisPage = amount
	} else {
		amountFromThisPage = length - startingPoint
	}

	log.Printf("amount from this page: %d", amountFromThisPage)
	fromThisPage := make([]Tangible, amountFromThisPage)
	var fromLaterPages []Tangible
	var nextCollection Container
	var nextStartingPoint uint

	// TODO: parallelize this

	for i := uint(0); i < amountFromThisPage; i++ {
		fromThisPage[i] = NewTangible(c.elements[i+startingPoint], c.id)
	}

	if errors.Is(c.nextErr, object.ErrKeyNotPresent) || length > amount + startingPoint {
		fromLaterPages, nextCollection, nextStartingPoint = []Tangible{}, c, amount + startingPoint
	} else {
		if c.nextErr != nil {
			fromLaterPages, nextCollection, nextStartingPoint = []Tangible{NewFailure(c.nextErr)}, c, amount + startingPoint
		} else if next, err := NewCollection(c.next, c.id); err != nil {
			fromLaterPages, nextCollection, nextStartingPoint = []Tangible{NewFailure(err)}, c, amount + startingPoint
		} else {
			fromLaterPages, nextCollection, nextStartingPoint = next.Harvest(amount - amountFromThisPage, 0)
		}
	}

	return append(fromThisPage, fromLaterPages...), nextCollection, nextStartingPoint
}
