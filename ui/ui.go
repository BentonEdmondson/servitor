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
)

type State struct {
	// TODO: the part stored in the history array is
	// called page, page will be renamed to children
	m *sync.Mutex

	feed  *feed.Feed
	index int

	frontier  pub.Tangible
	loadingUp bool

	page        pub.Container
	basepoint   uint
	loadingDown bool

	width  int
	height int
	output func(string)

	config *config.Config
}

func (s *State) view() string {
	if s.feed.IsEmpty() {
		return ansi.CenterVertically("", style.Color("  Loading…"), "", uint(s.height))
	}

	var top, center, bottom string
	for i := s.index - s.config.Context; i <= s.index+s.config.Context; i++ {
		if !s.feed.Contains(i) {
			continue
		}
		var serialized string
		if i == 0 {
			serialized = s.feed.Get(i).String(s.width - 4)
		} else if i > 0 {
			serialized = "→ " + ansi.Indent(s.feed.Get(i).Preview(s.width-4), "  ", false)
		} else {
			serialized = s.feed.Get(i).Preview(s.width - 4)
		}
		if i == s.index {
			center = ansi.Indent(serialized, "┃ ", true)
		} else if i < s.index {
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
	if s.loadingUp {
		if top != "" {
			top += "\n"
		}
		top = "  " + style.Color("Loading…") + "\n" + top
	}
	if s.loadingDown {
		if bottom != "" {
			bottom += "\n"
		}
		bottom += "\n  " + style.Color("Loading…")
	}
	return ansi.CenterVertically(top, center, bottom, uint(s.height))
}

func (s *State) Update(input byte) {
	switch input {
	case 'k': // up
		s.m.Lock()
		if s.feed.Contains(s.index - 1) {
			s.index -= 1
		}
		s.loadSurroundings()
		s.output(s.view())
		s.m.Unlock()
	case 'j': // down
		s.m.Lock()
		if s.feed.Contains(s.index + 1) {
			s.index += 1
		}
		s.loadSurroundings()
		s.output(s.view())
		s.m.Unlock()
	case 'g': // return to OP
		s.m.Lock()
		if s.feed.Contains(0) {
			s.index = 0
			s.output(s.view())
		}
		s.m.Unlock()
	case ' ': // select
		s.m.Lock()
		s.switchTo(s.feed.Get(s.index))
		s.m.Unlock()
	}
	// TODO: the catchall down here will be to look at s.feed.Get(s.index).References()
	// for urls to switch to
}

func (s *State) switchTo(item pub.Any) {
	s.loadingUp = false
	s.loadingDown = false
	s.basepoint = 0
	switch narrowed := item.(type) {
	case pub.Tangible:
		s.feed = feed.Create(narrowed)
		s.index = 0
		s.page = narrowed.Children()
		s.frontier = narrowed
	case pub.Container:
		s.feed = feed.CreateEmpty()
		s.index = 1
		s.page = narrowed
		s.frontier = nil
	default:
		panic(fmt.Sprintf("unrecognized non-Tangible non-Container: %T", item))
	}
	s.loadSurroundings()
	s.output(s.view())
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
	var prior State = *s
	if !s.loadingUp && !s.feed.Contains(s.index-s.config.Context) && s.frontier != nil {
		s.loadingUp = true
		go func() {
			parents, newFrontier := prior.frontier.Parents(uint(prior.config.Context))
			s.m.Lock()
			prior.feed.Prepend(parents)
			if prior.feed == s.feed {
				s.frontier = newFrontier
				s.loadingUp = false
				s.output(s.view())
			}
			s.m.Unlock()
		}()
	}
	if !s.loadingDown && !s.feed.Contains(s.index+s.config.Context) && s.page != nil {
		s.loadingDown = true
		go func() {
			children, newPage, newBasepoint := prior.page.Harvest(uint(prior.config.Context), prior.basepoint)
			s.m.Lock()
			prior.feed.Append(children)
			if prior.feed == s.feed {
				s.page = newPage
				s.basepoint = newBasepoint
				s.loadingDown = false
				s.output(s.view())
			}
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
		s.m.Unlock()
	}()
}

func NewState(config *config.Config, width int, height int, output func(string)) *State {
	s := &State{
		feed:   feed.CreateEmpty(),
		index:  0,
		config: config,
		width:  width,
		height: height,
		output: output,
		m:      &sync.Mutex{},
	}
	return s
}
