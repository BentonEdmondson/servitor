package pub

import (
	"mimicry/style"
	"time"
	"mimicry/mime"
)

type Failure struct {
	message error
}

func NewFailure(err error) *Failure {
	if err == nil {
		panic("can't create Failure with a nil error")
	}
	return &Failure{err}
}

func (f *Failure) Kind() string { return "failure" }

func (f *Failure) Name() string {
	return style.Problem(f.message)
}

func (f *Failure) Preview(width int) string {
	return f.Name()
}

func (f *Failure) String(width int) string {
	return f.Preview(width)
}

func (f *Failure) Parents(uint) ([]Tangible, Tangible) {
	return []Tangible{}, nil
}

func (f *Failure) Children() Container {
	return nil
}

func (f *Failure) Timestamp() time.Time {
	return time.Time{}
}

func (f *Failure) SelectLink(input int) (string, *mime.MediaType, bool) {
	return "", nil, false
}
