package history

type History[T any] struct {
	elements []T
	index    int
}

func (h *History[T]) Current() T {
	return h.elements[h.index]
}

func (h *History[T]) Back() {
	if h.index > 0 {
		h.index -= 1
	}
}

func (h *History[T]) Forward() {
	if len(h.elements) > h.index+1 {
		h.index += 1
	}
}

func (h *History[T]) Add(element T) {
	if h.elements == nil {
		h.elements = []T{element}
		h.index = 0
		return
	}
	h.elements = append(h.elements[:h.index+1], element)
	h.index += 1
}

func (h *History[T]) IsEmpty() bool {
	return len(h.elements) == 0
}
