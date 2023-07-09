package splicer

import (
	"servitor/pub"
	"sync"
)

type Splicer []struct {
	basepoint uint
	page      pub.Container
	elements   []pub.Tangible
}

func (s Splicer) Harvest(quantity uint, startingPoint uint) ([]pub.Tangible, pub.Container, uint) {
	/* Make a clone so Splicer remains immutable and thus threadsafe */
	clone := s.clone()

	clone.replenish(int(quantity + startingPoint))

	for i := 0; i < int(startingPoint); i++ {
		_ = clone.microharvest()
	}

	output := make([]pub.Tangible, 0, quantity)
	for i := 0; i < int(quantity); i++ {
		harvested := clone.microharvest()
		if harvested == nil {
			clone = nil
			break
		}
		output = append(output, harvested)
	}

	return output, clone, 0
}

func (s Splicer) clone() *Splicer {
	newSplicer := make(Splicer, len(s))
	copy(newSplicer, s)
	for i := range newSplicer {
		copy(newSplicer[i].elements, s[i].elements)
	}
	return &newSplicer
}

func (s Splicer) replenish(amount int) {
	var wg sync.WaitGroup
	for i, source := range s {
		i := i
		source := source
		wg.Add(1)
		go func() {
			if len(source.elements) < amount && source.page != nil {
				var newElements []pub.Tangible
				newElements, s[i].page, s[i].basepoint = source.page.Harvest(uint(amount - len(source.elements)), source.basepoint)
				s[i].elements = append(s[i].elements, newElements...)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func (s Splicer) microharvest() pub.Tangible {
	var mostRecent pub.Tangible
	var mostRecentIndex int
	for i, candidate := range s {
		if len(candidate.elements) == 0 {
			continue
		}

		candidateElement := candidate.elements[0]

		if mostRecent == nil {
			mostRecent = candidateElement
			mostRecentIndex = i
			continue
		}

		if candidateElement == nil {
			continue
		}

		if candidateElement.Timestamp().After(mostRecent.Timestamp()) {
			mostRecent = candidateElement
			mostRecentIndex = i
			continue
		}
	}

	if mostRecent == nil {
		return nil
	}

	/* Shift (pop from front) the element that was selected */
	s[mostRecentIndex].elements = s[mostRecentIndex].elements[1:]

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
			switch narrowed := fetched.(type) {
			case pub.Tangible:
				s[i].page = narrowed.Children()
			case *pub.Collection:
				s[i].page = narrowed
			default:
				panic("cannot splice non-Tangible, non-Collection")
			}
			s[i].basepoint = 0
			s[i].elements = []pub.Tangible{}
			wg.Done()
		}()
	}
	wg.Wait()

	return &s
}
