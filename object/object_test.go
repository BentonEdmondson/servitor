package object

import (
	"errors"
	"testing"
)

func TestString(t *testing.T) {
	o := Object{
		"good": "value",
		"bad":  float64(25),
		// deliberately absent: "absent": "value",
	}
	str, err := o.GetString("good")
	if err != nil {
		t.Fatalf("Problem extracting string: %v", err)
	}
	if str != "value" {
		t.Fatalf(`Expected "value" not %v`, str)
	}

	_, err = o.GetString("bad")
	if !errors.Is(err, ErrKeyWrongType) {
		t.Fatalf(`Expected ErrKeyWrongType, not %v`, err)
	}

	_, err = o.GetString("absent")
	if !errors.Is(err, ErrKeyNotPresent) {
		t.Fatalf(`Expected ErrKeyNotPresent, not %v`, err)
	}
}

func TestNumber(t *testing.T) {
	o := Object{
		"good": float64(25),
		"bad":  "value",
		// deliberately absent: "absent": "value",
	}
	num, err := o.GetNumber("good")
	if err != nil {
		t.Fatalf("Problem extracting number: %v", err)
	}
	if num != 25 {
		t.Fatalf(`Expected 25 not %v`, num)
	}

	_, err = o.GetNumber("bad")
	if !errors.Is(err, ErrKeyWrongType) {
		t.Fatalf(`Expected ErrKeyWrongType, not %v`, err)
	}

	_, err = o.GetNumber("absent")
	if !errors.Is(err, ErrKeyNotPresent) {
		t.Fatalf(`Expected ErrKeyNotPresent, not %v`, err)
	}
}

func TestObject(t *testing.T) {
	o := Object{
		"good": map[string]any{},
		"bad":  "value",
		// deliberately absent: "absent": "value",
	}
	obj, err := o.GetObject("good")
	if err != nil {
		t.Fatalf("Problem extracting Object: %v", err)
	}
	if len(obj) != 0 {
		t.Fatalf(`Expected empty map, not %v`, obj)
	}

	_, err = o.GetObject("bad")
	if !errors.Is(err, ErrKeyWrongType) {
		t.Fatalf(`Expected ErrKeyWrongType, not %v`, err)
	}

	_, err = o.GetObject("absent")
	if !errors.Is(err, ErrKeyNotPresent) {
		t.Fatalf(`Expected ErrKeyNotPresent, not %v`, err)
	}
}

func TestList(t *testing.T) {
	o := Object{
		"multiple": []any{"first", "second"},
		"single":   "one",
		// deliberately absent: "absent": "value",
	}
	list, err := o.GetList("multiple")
	if err != nil {
		t.Fatalf("Problem extracting list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf(`Expected 2 elements, but didn't get them: %v`, list)
	}

	list, err = o.GetList("single")
	if err != nil {
		t.Fatalf("Problem extracting list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf(`Expected 1 element to auto-convert to list, but didn't: %v`, list)
	}

	_, err = o.GetList("absent")
	if !errors.Is(err, ErrKeyNotPresent) {
		t.Fatalf(`Expected ErrKeyNotPresent, not %v`, err)
	}
}
