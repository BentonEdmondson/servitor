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
	"os/exec"
	"mimicry/mime"
	"strings"
	"strconv"
)

/*
	The public methods herein are threadsafe, the private methods
	are not and need to be protected by State.m
*/

/* Modes */
const (
	loading = iota
	normal
	command
	selection
	opening
	problem
)

const (
	enterKey byte = '\r'
	escapeKey byte = 27
	backspaceKey byte = 127
)

type Page struct {
	feed  *feed.Feed

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

	mode int
	buffer string
}

func (s *State) view() string {
	if s.mode == loading {
		return ansi.CenterVertically("", style.Color("  Loading…"), "", uint(s.height))
	}

	var top, center, bottom string
	for i := -s.config.Context; i <= s.config.Context; i++ {
		if !s.h.Current().feed.Contains(i) {
			continue
		}
		var serialized string
		if s.h.Current().feed.Location(i) == 0 {
			serialized = s.h.Current().feed.Get(i).String(s.width - 4)
		} else if s.h.Current().feed.Location(i) > 0 {
			serialized = "→ " + ansi.Indent(s.h.Current().feed.Get(i).Preview(s.width-8), "  ", false)
		} else {
			serialized = s.h.Current().feed.Get(i).Preview(s.width - 4)
		}
		if i == 0 {
			center = ansi.Indent(serialized, "┃ ", true)
		} else if i < 0 {
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
	output := ansi.CenterVertically(top, center, bottom, uint(s.height))
	
	var footer string
	switch s.mode {
	case normal:
		break
	case selection:
		footer = "Selecting " + s.buffer + " (press . to open internally, enter to open externally)"
	case command:
		footer = ":" + s.buffer
	case opening:
		footer = "Opening " + s.buffer + "\u2026"
	case problem:
		footer = s.buffer
	default:
		panic("encountered unrecognized mode")
	}
	if footer != "" {
		output = ansi.ReplaceLastLine(output, style.Highlight(ansi.SetLength(footer, s.width, "\u2026")))
	}

	return output
}

func (s *State) Update(input byte) {
	s.m.Lock()
	defer s.m.Unlock()

	if s.mode == loading {
		panic("inputs should not come through while loading, as all loading functions block the UI thread")
	}

	if input == escapeKey {
		s.buffer = ""
		s.mode = normal
		s.output(s.view())
		return
	}

	if input == backspaceKey {
		if len(s.buffer) == 0 {
			s.mode = normal
			s.output(s.view())
			return
		}
		s.buffer = s.buffer[:len(s.buffer)-1]
		if s.buffer == "" && s.mode == selection {
			s.mode = normal
		}
		s.output(s.view())
		return
	}

	if s.mode == command {
		if input == enterKey {
			if args := strings.SplitN(s.buffer, " ", 2); len(args) == 2 {
				err := s.subcommand(args[0], args[1])
				if err != nil {
					s.buffer = "Failed to run command: " + ansi.Squash(err.Error())
					s.mode = problem
					s.output(s.view())
					s.buffer = ""
					s.mode = normal
				}
			} else {
				s.buffer = ""
				s.mode = normal
			}
			return
		}
		s.buffer += string(input)
		s.output(s.view())
		return
	}

	if input == ':' {
		s.buffer = ""
		s.mode = command
		s.output(s.view())
		return
	}

	if input >= '0' && input <= '9' {
		if s.mode != selection {
			s.buffer = ""
		}
		s.buffer += string(input)
		s.mode = selection
		s.output(s.view())
		return
	}

	if s.mode == selection {
		if input == '.' || input == enterKey {
			number, err := strconv.Atoi(s.buffer)
			if err != nil {
				panic("buffer had a non-number while in selection mode")
			}
			link, mediaType, present := s.h.Current().feed.Current().SelectLink(number)
			if !present {
				s.buffer = ""
				s.mode = normal
				s.output(s.view())
				return
			}
			if input == '.' {
				s.openInternally(link)
			}
			if input == enterKey {
				s.openExternally(link, mediaType)
			}
			return
		}
		/* At this point we know input is a non-number, non-., non-enter */
		s.mode = normal
		s.buffer = ""
	}

	/* At this point we know we are in normal mode */
	switch input {
	// TODO: make feed stateful so all this logic is nicer. Functions will be MoveUp, MoveDown, MoveToCenter
	case 'k': // up
		s.h.Current().feed.MoveUp()
		s.loadSurroundings()
	case 'j': // down
		s.h.Current().feed.MoveDown()
		s.loadSurroundings()
	case 'g': // return to OP
		s.h.Current().feed.MoveToCenter()
	case 'h': // back in history
		s.h.Back()
	case 'l': // forward in history
		s.h.Forward()
	case ' ': // select
		s.switchTo(s.h.Current().feed.Current())
	case 'c': // get creator of post
		unwrapped := s.h.Current().feed.Current()
		if activity, ok := unwrapped.(*pub.Activity); ok {
			unwrapped = activity.Target()
		}
		if post, ok := unwrapped.(*pub.Post); ok {
			creators := post.Creators()
			s.switchTo(creators)
		}
	case 'r': // get recipient of post
		unwrapped := s.h.Current().feed.Current()
		if activity, ok := unwrapped.(*pub.Activity); ok {
			unwrapped = activity.Target()
		}
		if post, ok := unwrapped.(*pub.Post); ok {
			recipients := post.Recipients()
			s.switchTo(recipients)
		}
	case 'a': // get actor of activity
		if activity, ok := s.h.Current().feed.Current().(*pub.Activity); ok {
			actor := activity.Actor()
			s.switchTo(actor)
		}
	case 'o':
		unwrapped := s.h.Current().feed.Current()
		if activity, ok := unwrapped.(*pub.Activity); ok {
			unwrapped = activity.Target()
		}
		if post, ok := unwrapped.(*pub.Post); ok {
			if link, mediaType, present := post.Media(); present {
				s.openExternally(link, mediaType)
			}
		}
	case 'p':
		if actor, ok := s.h.Current().feed.Current().(*pub.Actor); ok {
			if link, mediaType, present := actor.ProfilePic(); present {
				s.openExternally(link, mediaType)
			}
		}
	case 'b':
		if actor, ok := s.h.Current().feed.Current().(*pub.Actor); ok {
			if link, mediaType, present := actor.Banner(); present {
				s.openExternally(link, mediaType)
			}
		}
	}
	s.output(s.view())
}

func (s *State) switchTo(item any) {
	switch narrowed := item.(type) {
	case []pub.Tangible:
		if len(narrowed) == 0 {
			return
		}
		if len(narrowed) == 1 {
			s.h.Add(&Page{
				feed: feed.Create(narrowed[0]),
				children: narrowed[0].Children(),
				frontier: narrowed[0],
			})
		} else {
			s.h.Add(&Page{
				feed: feed.CreateAndAppend(narrowed),
			})
		}
	case pub.Tangible:
		s.h.Add(&Page{
			feed: feed.Create(narrowed),
			children: narrowed.Children(),
			frontier: narrowed,
		})
	case pub.Container:
		s.mode = loading
		s.buffer = ""
		s.output(s.view())
		children, nextCollection, newBasepoint := narrowed.Harvest(uint(s.config.Context), 0)
		s.h.Add(&Page{
			basepoint: newBasepoint,
			children: nextCollection,
			feed: feed.CreateAndAppend(children),
		})
	default:
		panic(fmt.Sprintf("unrecognized non-Tangible non-Container: %T", item))
	}
	s.mode = normal
	s.buffer = ""
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
	if !page.loadingUp && !page.feed.Contains(-context) && page.frontier != nil {
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
	if !page.loadingDown && !page.feed.Contains(context) && page.children != nil {
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

func (s *State) openUserInput(input string) {
	s.mode = loading
	s.buffer = ""
	s.output(s.view())
	result := pub.FetchUserInput(input)
	s.switchTo(result)
	s.output(s.view())
}

func (s *State) openInternally(input string) {
	s.mode = loading
	s.buffer = ""
	s.output(s.view())
	result := pub.New(input, nil)
	s.switchTo(result)
	s.output(s.view())
}

func (s *State) openFeed(input string) {
	inputs, present := s.config.Feeds[input]
	if !present {
		s.mode = problem
		s.buffer = "Failed to open feed: " + input + " is not a known feed"
		s.output(s.view())
		s.mode = normal
		s.buffer = ""
		return
	}
	s.mode = loading
	s.buffer = ""
	s.output(s.view())
	result := splicer.NewSplicer(inputs)
	s.switchTo(result)
	s.output(s.view())
}

func NewState(config *config.Config, width int, height int, output func(string)) *State {
	s := &State{
		h: history.History[*Page]{},
		config: config,
		width:  width,
		height: height,
		output: output,
		m:      &sync.Mutex{},
		mode:   loading,
	}
	return s
}

func (s *State) Subcommand(name, argument string) error {
	s.m.Lock()
	defer s.m.Unlock()
	return s.subcommand(name, argument)
}

func (s *State) subcommand(name, argument string) error {
	switch name {
	case "open":
		s.openUserInput(argument)
	case "feed":
		s.openFeed(argument)
	default:
		return fmt.Errorf("unrecognized subcommand: %s", name)
	}
	return nil
}

func (s *State) openExternally(link string, mediaType *mime.MediaType) {
	s.mode = opening
	s.buffer = link
	s.output(s.view())

	command := make([]string, len(s.config.MediaHook))
	copy(command, s.config.MediaHook)

	foundPercentU := false
	for i, field := range command {
		if i == 0 {
			continue
		}
		switch field {
		case "%u":
			command[i] = link
			foundPercentU = true
		case "%m":
			command[i] = mediaType.Essence
		case "%s":
			command[i] = mediaType.Subtype
		case "%t":
			command[i] = mediaType.Supertype
		}
	}

	cmd := exec.Command(command[0], command[1:]...)
	if !foundPercentU {
		cmd.Stdin = strings.NewReader(link)
	}

	go func() {
		err := cmd.Run()

		s.m.Lock()
		defer s.m.Unlock()
		
		if s.mode != opening {
			return
		}

		if err != nil {
			s.mode = problem
			s.buffer = "Failed to open link: " + ansi.Squash(err.Error())
			s.output(s.view())
			s.mode = normal
			s.buffer = ""
			return
		}
	
		s.mode = normal
		s.buffer = ""
		s.output(s.view())
	}()
}
