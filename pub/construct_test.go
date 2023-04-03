package pub

import (
	"mimicry/util"
	"testing"
)

func TestFromFile(t *testing.T) {
	item, err := FetchUserInput("../tests/cases/basic.json")
	if err != nil { t.Fatal(err) }
	note, ok := item.(Post)
	if !ok { t.Fatal("basic.json is not a Post") }

	util.AssertEqual("Note", note.Kind(), t)
	body, err := note.Body(100)
	if err != nil { t.Fatal(err) }
	util.AssertEqual("Hello, World!", body, t)
}