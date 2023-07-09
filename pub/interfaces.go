package pub

import (
	"servitor/mime"
	"time"
)

type Any any

type Tangible interface {
	String(width int) string
	Preview(width int) string
	Parents(quantity uint) ([]Tangible, Tangible)
	Children() Container
	Timestamp() time.Time
	Name() string
	SelectLink(input int) (string, *mime.MediaType, bool)
}

type Container interface {
	/* result, index of next item, next collection */
	Harvest(quantity uint, startingAt uint) ([]Tangible, Container, uint)
}
