package kinds

type Content interface {
	String() string
	Kind() (string, error)
	Category() string
}