package feed

import (
	"fmt"
	"mimicry/pub"
)

type Feed struct {
	feed       map[int]pub.Tangible
	upperBound int
	lowerBound int
}

func Create(input pub.Tangible) *Feed {
	return &Feed{
		feed: map[int]pub.Tangible{
			0: input,
		},
		upperBound: 0,
		lowerBound: 0,
	}
}

func CreateAndAppend(input []pub.Tangible) *Feed {
	f := &Feed{
		feed: map[int]pub.Tangible{},
	}
	f.Append(input)
	f.lowerBound = 1
	return f
}

func (f *Feed) Append(input []pub.Tangible) {
	for i, element := range input {
		f.feed[f.upperBound+i+1] = element
	}
	f.upperBound += len(input)
}

func (f *Feed) Prepend(input []pub.Tangible) {
	for i, element := range input {
		f.feed[f.lowerBound-i-1] = element
	}
	f.lowerBound -= len(input)
}

func (f *Feed) Get(index int) pub.Tangible {
	if index > f.upperBound || index < f.lowerBound {
		panic(fmt.Sprintf("indexing feed at %d whereas bounds are %d and %d", index, f.lowerBound, f.upperBound))
	}

	return f.feed[index]
}

func (f *Feed) Contains(index int) bool {
	return index <= f.upperBound && index >= f.lowerBound
}
