package history

import (
	"testing"
)

func TestCurrent(t *testing.T) {
	h := History[int]{}
	h.Add(1)
	current := h.Current()
	if current != 1 {
		t.Fatalf("current should be 1 but is %v", current)
	}
}

func TestBackForward(t *testing.T) {
	h := History[int]{}
	h.Add(1)
	h.Add(2)
	h.Back()
	back := h.Current()
	h.Forward()
	forward := h.Current()
	if back != 1 || forward != 2 {
		t.Fatalf("back should be 1 not %v, forward should be 2 not %v", back, forward)
	}
}

func TestIsEmpty(t *testing.T) {
	h := History[int]{}
	if !h.IsEmpty() {
		t.Fatalf("history should report empty when empty")
	}
	h.Add(1)
	if h.IsEmpty() {
		t.Fatalf("history is purporting to be empty when it shouldn't be")
	}
}

func TestBackSaturation(t *testing.T) {
	h := History[int]{}
	h.Add(1)
	h.Add(2)
	h.Back()
	h.Back()
	h.Back()
	current := h.Current()
	if current != 1 {
		t.Fatalf("current should be 1 not %v after back saturation", current)
	}
}

func TestForwardSaturation(t *testing.T) {
	h := History[int]{}
	h.Add(1)
	h.Add(2)
	h.Forward()
	h.Forward()
	current := h.Current()
	if current != 2 {
		t.Fatalf("current should be 2 not %v after forward saturation", current)
	}
}

func TestForwardDestruction(t *testing.T) {
	h := History[int]{}
	h.Add(1)
	h.Add(2)
	h.Back()
	h.Add(3)
	h.Forward()
	h.Forward()
	current := h.Current()
	if current != 3 {
		t.Fatalf("current should be 3 not %v after forward destruction", current)
	}
}
