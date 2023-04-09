package mime

import (
	"testing"
)

func TestDefault(t *testing.T) {
	m := Default()
	if m.Essence != "text/html" {
		t.Fatalf(`Default media type should be "text/html", not %#v`, m.Essence)
	}
	if m.Supertype != "text" {
		t.Fatalf(`Default media type supertype should be "text", not %#v`, m.Supertype)
	}
	if m.Subtype != "html" {
		t.Fatalf(`Default media type subtype should be "html", not %#v`, m.Subtype)
	}
}

func TestFailedParse(t *testing.T) {
	m, err := Parse("")
	if err == nil {
		t.Fatalf("Should fail to parse an empty string, but instead returns: %#v", m)
	}

	m, err = Parse("application")
	if err == nil {
		t.Fatalf("Should fail to parse invalid media type, but instead returns: %#v", m)
	}
}

func TestSuccessfulUpdate(t *testing.T) {
	m := Default()
	err := m.Update("application/json ; charset=utf-8")
	if err != nil {
		t.Fatalf("Update should have succeeded but returned error: %#v", err)
	}
	if m.Essence != "application/json" {
		t.Fatalf(`New media type should be "application/json", not %#v`, m.Essence)
	}
	if m.Supertype != "application" {
		t.Fatalf(`New media type supertype should be "application", not %#v`, m.Supertype)
	}
	if m.Subtype != "json" {
		t.Fatalf(`New media type subtype should be "json", not %#v`, m.Subtype)
	}
}

func TestFailedUpdate(t *testing.T) {
	m := Default()
	err := m.Update("no slash")
	if err == nil {
		t.Fatalf(`Expected "no slash" to result in an Update error, but it resulted in: %#v`, m)
	}
}

func TestMatchesSuccess(t *testing.T) {
	m := Default()
	matches := m.Matches([]string{"application/json", "text/html"})
	if !matches {
		t.Fatalf(`Expected media type to match text/html but it did not: %#v`, m)
	}
}

func TestMatchesFailure(t *testing.T) {
	m := Default()
	matches := m.Matches([]string{"application/json"})
	if matches {
		t.Fatalf(`Expected media type to not match application/json: %#v`, m)
	}
}
