package ui

import (
	"fmt"
	"mimicry/ansi"
	"mimicry/config"
	"mimicry/feed"
	"mimicry/pub"
	"mimicry/splicer"
	"mimicry/style"
	"sync"
	"mimicry/history"
)

/*
	The public methods herein are threadsafe, the private methods
	are not and need to be protected by State.m
*/

type Page struct {
	feed  *feed.Feed
	index int

	frontier  pub.Tangible
	loadingUp bool

	children        pub.Container
	basepoint   uint
	loadingDown bool
}

type State struct {
	m *sync.Mutex

	h history.History[*Page]

	width  int
	height int
	output func(string)

	config *config.Config
}

func (s *State) view() string {
	if s.h.IsEmpty() || s.h.Current().feed.IsEmpty() {
		return ansi.CenterVertically("", style.Color("  Loading…"), "", uint(s.height))
	}

	var top, center, bottom string
	for i := s.h.Current().index - s.config.Context; i <= s.h.Current().index+s.config.Context; i++ {
		if !s.h.Current().feed.Contains(i) {
			continue
		}
		var serialized string
		if i == 0 {
			serialized = s.h.Current().feed.Get(i).String(s.width - 4)
		} else if i > 0 {
			serialized = "→ " + ansi.Indent(s.h.Current().feed.Get(i).Preview(s.width-4), "  ", false)
		} else {
			serialized = s.h.Current().feed.Get(i).Preview(s.width - 4)
		}
		if i == s.h.Current().index {
			center = ansi.Indent(serialized, "┃ ", true)
		} else if i < s.h.Current().index {
			if top != "" {
				top += "\n"
			}
			top += ansi.Indent(serialized+"\n", "  ", true)
		} else {
			if bottom != "" {
				bottom += "\n"
			}
			bottom += ansi.Indent("\n"+serialized, "  ", true)
		}
	}
	if s.h.Current().loadingUp {
		if top != "" {
			top += "\n"
		}
		top = "  " + style.Color("Loading…") + "\n" + top
	}
	if s.h.Current().loadingDown {
		if bottom != "" {
			bottom += "\n"
		}
		bottom += "\n  " + style.Color("Loading…")
	}
	return ansi.CenterVertically(top, center, bottom, uint(s.height))
}

func (s *State) Update(input byte) {
	s.m.Lock()
	defer s.m.Unlock()
	switch input {
	case 'k': // up
		if s.h.Current().feed.Contains(s.h.Current().index - 1) {
			s.h.Current().index -= 1
		}
		s.output(s.view())
		s.loadSurroundings()
	case 'j': // down
		if s.h.Current().feed.Contains(s.h.Current().index + 1) {
			s.h.Current().index += 1
		}
		s.output(s.view())
		s.loadSurroundings()
	case 'g': // return to OP
		if s.h.Current().feed.Contains(0) {
			s.h.Current().index = 0
		}
		s.output(s.view())
	case 'h': // back in history
		s.h.Back()
		s.output(s.view())
	case 'l':
		s.h.Forward()
		s.output(s.view())
	case ' ': // select
		s.switchTo(s.h.Current().feed.Get(s.h.Current().index))
		s.output(s.view())
	}
	// TODO: the catchall down here will be to look at s.feed.Get(s.index).References()
	// for urls to switch to
}

func (s *State) switchTo(item pub.Any) {
	switch narrowed := item.(type) {
	case pub.Tangible:
		s.h.Add(&Page{
			feed: feed.Create(narrowed),
			children: narrowed.Children(),
			frontier: narrowed,
		})
	case pub.Container:
		s.h.Add(&Page{
			feed: feed.CreateEmpty(),
			children: narrowed,
			index: 1,
		})
	default:
		panic(fmt.Sprintf("unrecognized non-Tangible non-Container: %T", item))
	}
	s.loadSurroundings()
}

func (s *State) SetWidthHeight(width int, height int) {
	s.m.Lock()
	defer s.m.Unlock()
	if s.width == width && s.height == height {
		return
	}
	s.width = width
	s.height = height
	s.output(s.view())
}

func (s *State) loadSurroundings() {
	page := s.h.Current()
	context := s.config.Context
	if !page.loadingUp && !page.feed.Contains(page.index-context) && page.frontier != nil {
		page.loadingUp = true
		go func() {
			parents, newFrontier := page.frontier.Parents(uint(context))
			s.m.Lock()
			page.feed.Prepend(parents)
			page.frontier = newFrontier
			page.loadingUp = false
			s.output(s.view())
			s.m.Unlock()
		}()
	}
	if !page.loadingDown && !page.feed.Contains(page.index+context) && page.children != nil {
		page.loadingDown = true
		go func() {
			// TODO: need to do a new renaming, maybe upperFrontier, lowerFrontier
			children, nextCollection, newBasepoint := page.children.Harvest(uint(context), page.basepoint)
			s.m.Lock()
			page.feed.Append(children)
			page.children = nextCollection
			page.basepoint = newBasepoint
			page.loadingDown = false
			s.output(s.view())
			s.m.Unlock()
		}()
	}
}

func (s *State) Open(input string) {
	go func() {
		s.m.Lock()
		s.output(s.view())
		s.m.Unlock()
		result := pub.FetchUserInput(input)
		s.m.Lock()
		s.switchTo(result)
		s.output(s.view())
		s.m.Unlock()
	}()
}

func (s *State) Feed(input string) {
	go func() {
		s.m.Lock()
		s.output(s.view())
		inputs := s.config.Feeds[input]
		s.m.Unlock()
		result := splicer.NewSplicer(inputs)
		s.m.Lock()
		s.switchTo(result)
		s.output(s.view())
		s.m.Unlock()
	}()
}

func NewState(config *config.Config, width int, height int, output func(string)) *State {
	s := &State{
		h: history.History[*Page]{},
		config: config,
		width:  width,
		height: height,
		output: output,
		m:      &sync.Mutex{},
	}
	return s
}
