package ui

import (
	"mimicry/pub"
	"mimicry/ansi"
	"mimicry/feed"
	"fmt"
	"log"
)

type State struct {
	/* the 0 index is special; it is rendered in full, not as a preview.
	   the others are all rendered as previews. negative indexes represent
	   parents of the 0th element (found via `inReplyTo`) and positive
	   elements represent children (found via the pertinent collection,
	   e.g. `replies` for a post or `outbox` for an actor) */
	feed *feed.Feed
	index int
	context int

	page pub.Container
	basepoint uint
}

func (s *State) View(width int, height uint) string {
	//return s.feed.Get(0).String(width)
	var top, center, bottom string
	//TODO: this should be bounded based on size of feed
	for i := s.index - s.context; i <= s.index + s.context; i++ {
		if !s.feed.Contains(i) {
			continue
		}
		var serialized string
		if i == 0 {
			serialized = s.feed.Get(i).String(width - 4)
			log.Printf("%d\n", len(serialized))
		} else if i > 0 {
			serialized = "╰ " + ansi.Indent(s.feed.Get(i).Preview(width - 4), "  ", false)
		} else {
			serialized = s.feed.Get(i).Preview(width - 4)
		}
		if i == s.index {
			center = ansi.Indent(serialized, "┃ ", true)
		} else if i < s.index {
			if top != "" { top += "\n" }
			top += ansi.Indent(serialized + "\n│", "  ", true)
		} else {
			if bottom != "" { bottom = "\n" + bottom }
			bottom = ansi.Indent("│\n" + serialized, "  ", true) + bottom
		}
	}
	log.Printf("%s\n", center)
	return ansi.CenterVertically(top, center, bottom, height)
}

func (s *State) Update(input byte) {
	/* Interesting problem, but you will succeed! */
	switch input {
	case 'k': // up
		mayNeedLoading := s.index - 1 - s.context
		if !s.feed.Contains(mayNeedLoading) {
			if s.feed.Contains(mayNeedLoading - 1) {
				s.feed.Prepend(s.feed.Get(mayNeedLoading - 1).Parents(1))
			}
		}

		if s.feed.Contains(s.index - 1) {
			s.index -= 1
		}
	case 'j': // down
		mayNeedLoading := s.index + 1 + s.context
		if !s.feed.Contains(mayNeedLoading) {
			if s.page != nil {
				var children []pub.Tangible
				children, s.page, s.basepoint = s.page.Harvest(1, s.basepoint)
				s.feed.Append(children)
			}
		}

		if s.feed.Contains(s.index + 1) {
			s.index += 1
		}
	}
	// TODO: the catchall down here will be to look at s.feed.Get(s.index).References()
	// for urls to switch to
}

func (s *State) SwitchTo(item pub.Any)  {
	switch narrowed := item.(type) {
	case pub.Tangible:
		s.feed = feed.Create(narrowed)
		s.feed.Prepend(narrowed.Parents(uint(s.context)))
		var children []pub.Tangible
		children, s.page, s.basepoint = narrowed.Children(uint(s.context))
		s.feed.Append(children)
	case pub.Container:
		var children []pub.Tangible
		children, s.page, s.basepoint = narrowed.Harvest(uint(s.context), 0)
		s.feed = feed.CreateAndAppend(children)
	default:
		panic(fmt.Sprintf("unrecognized non-Tangible non-Container: %T", item))
	}
}

func Start(input string) *State {
	item := pub.FetchUserInput(input)
	log.Printf("%v\n", item)
	s := &State{
		feed: &feed.Feed{},
		index: 0,
		context: 1,
	}
	s.SwitchTo(item)
	return s
}
