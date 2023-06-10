package feed

import (
	"fmt"
	"mimicry/pub"
)

type Feed struct {
	feed map[int]pub.Tangible

	// exclusive bounds
	upperBound int
	lowerBound int
}

func CreateEmpty() *Feed {
	return &Feed{
		feed:       map[int]pub.Tangible{},
		upperBound: 0,
		lowerBound: 0,
	}
}

func Create(input pub.Tangible) *Feed {
	return &Feed{
		feed: map[int]pub.Tangible{
			0: input,
		},
		upperBound: 1,
		lowerBound: -1,
	}
}

func CreateAndAppend(input []pub.Tangible) *Feed {
	f := &Feed{
		feed: map[int]pub.Tangible{},
	}
	f.upperBound = 1
	f.Append(input)
	return f
}

func (f *Feed) Append(input []pub.Tangible) {
	for i, element := range input {
		f.feed[f.upperBound+i] = element
	}
	f.upperBound += len(input)
}

func (f *Feed) Prepend(input []pub.Tangible) {
	for i, element := range input {
		f.feed[f.lowerBound-i] = element
	}
	f.lowerBound -= len(input)
}

func (f *Feed) Get(index int) pub.Tangible {
	if !f.Contains(index) {
		panic(fmt.Sprintf("indexing feed at %d whereas bounds are %d and %d", index, f.lowerBound, f.upperBound))
	}

	return f.feed[index]
}

func (f *Feed) Contains(index int) bool {
	return index < f.upperBound && index > f.lowerBound
}

func (f *Feed) IsEmpty() bool {
	return f.upperBound == 0 && f.lowerBound == 0
}
