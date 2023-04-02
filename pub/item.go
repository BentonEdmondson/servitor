package pub

type Item interface {
	String(width int) (string, error)
	Preview() (string, error)
	Kind() string
	Category() string
}