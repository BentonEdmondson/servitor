package pub

type Any interface {
	Kind() string
}

type Tangible interface {
	Kind() string

	String(width int) string
	Preview(width int) string
	Parents(quantity uint) ([]Tangible, Tangible)
	Children() Container
}

type Container interface {
	Kind() string

	/* result, index of next item, next collection */
	Harvest(quantity uint, startingAt uint) ([]Tangible, Container, uint)
	Size() (uint64, error)
}