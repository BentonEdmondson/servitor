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

	index int
}

func CreateEmpty() *Feed {
	return &Feed{
		feed:       map[int]pub.Tangible{},
		upperBound: 0,
		lowerBound: 0,
		index:      0,
	}
}

func Create(input pub.Tangible) *Feed {
	return &Feed{
		feed: map[int]pub.Tangible{
			0: input,
		},
		upperBound: 1,
		lowerBound: -1,
		index:      0,
	}
}

func CreateAndAppend(input []pub.Tangible) *Feed {
	f := &Feed{
		feed: map[int]pub.Tangible{},
	}
	f.upperBound = 1
	f.index = 1
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

func (f *Feed) Get(offset int) pub.Tangible {
	if !f.Contains(offset) {
		panic(fmt.Sprintf("indexing feed at %d whereas bounds are %d and %d", f.index+offset, f.lowerBound, f.upperBound))
	}

	return f.feed[f.index+offset]
}

func (f *Feed) Current() pub.Tangible {
	return f.feed[f.index]
}

func (f *Feed) MoveUp() {
	if f.Contains(-1) {
		f.index -= 1
	}
}

func (f *Feed) MoveDown() {
	if f.Contains(1) {
		f.index += 1
	}
}

func (f *Feed) MoveToCenter() {
	if f.Contains(-f.index) {
		f.index = 0
	}
}

func (f *Feed) Contains(offset int) bool {
	return f.index+offset < f.upperBound && f.index+offset > f.lowerBound
}

func (f *Feed) IsParent(offset int) bool {
	return f.index+offset < 0
}

func (f *Feed) IsChild(offset int) bool {
	return f.index+offset > 0
}
