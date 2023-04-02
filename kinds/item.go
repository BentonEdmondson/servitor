package kinds

type Item interface {
	String(width int) (string, error)
	Preview() (string, error)
	Kind() string
	Category() string
}