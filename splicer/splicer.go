package splicer

import (
	"mimicry/pub"
	"sync"
)

type Splicer []struct {
	basepoint uint
	page      pub.Container
	element   pub.Tangible
}

func (s Splicer) Harvest(quantity uint, startingPoint uint) ([]pub.Tangible, pub.Container, uint) {
	clone := s.clone()

	for i := 0; i < int(startingPoint); i++ {
		clone.microharvest()
	}

	output := make([]pub.Tangible, 0, quantity)
	for i := 0; i < int(quantity); i++ {
		harvested := clone.microharvest()
		if harvested == nil {
			break
		}
		output = append(output, harvested)
	}

	return output, clone, 0
}

func (s Splicer) clone() *Splicer {
	newSplicer := make(Splicer, len(s))
	copy(newSplicer, s)
	return &newSplicer
}

func (s Splicer) microharvest() pub.Tangible {
	var mostRecent pub.Tangible
	var mostRecentIndex int
	for i, candidate := range s {
		if mostRecent == nil {
			mostRecent = candidate.element
			mostRecentIndex = i
			continue
		}

		if candidate.element == nil {
			continue
		}

		if candidate.element.Timestamp().After(mostRecent.Timestamp()) {
			mostRecent = candidate.element
			mostRecentIndex = i
			continue
		}
	}

	if mostRecent == nil {
		return nil
	}

	if s[mostRecentIndex].page != nil {
		var elements []pub.Tangible
		elements, s[mostRecentIndex].page, s[mostRecentIndex].basepoint = s[mostRecentIndex].page.Harvest(1, s[mostRecentIndex].basepoint)
		if len(elements) > 1 {
			panic("harvest returned more that one element when I only asked for one")
		} else {
			s[mostRecentIndex].element = elements[0]
		}
	} else {
		s[mostRecentIndex].element = nil
	}

	return mostRecent
}

func NewSplicer(inputs []string) *Splicer {
	s := make(Splicer, len(inputs))
	var wg sync.WaitGroup
	for i, input := range inputs {
		i := i
		input := input
		wg.Add(1)
		go func() {
			fetched := pub.FetchUserInput(input)
			var children pub.Container
			switch narrowed := fetched.(type) {
			case pub.Tangible:
				children = narrowed.Children()
			case *pub.Collection:
				children = narrowed
			default:
				panic("cannot splice non-Tangible, non-Collection")
			}

			if children != nil {
				var elements []pub.Tangible
				elements, s[i].page, s[i].basepoint = children.Harvest(1, 0)
				if len(elements) > 1 {
					panic("harvest returned more that one element when I only asked for one")
				} else {
					s[i].element = elements[0]
				}
			} else {
				s[i].element = nil
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return &s
}
