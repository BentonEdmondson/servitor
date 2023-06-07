package pub

import (
	"time"
)

type Any any

type Tangible interface {
	Kind() string

	String(width int) string
	Preview(width int) string
	Parents(quantity uint) ([]Tangible, Tangible)
	Children() Container
	Timestamp() time.Time
}

type Container interface {
	/* result, index of next item, next collection */
	Harvest(quantity uint, startingAt uint) ([]Tangible, Container, uint)
}