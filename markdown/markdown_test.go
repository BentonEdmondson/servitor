package markdown

import (
	"mimicry/style"
	"testing"
)

func TestBasic(t *testing.T) {
	input := `[Here's a link!](https://wikipedia.org)

![This is a beautiful image!](https://miro.medium.com/v2/resize:fit:900/0*L31Zh4YhAv3Wokco)

* Nested list
  * Nesting`
	markup, links, err := NewMarkup(input)
	if err != nil {
		t.Fatal(err)
	}

	if links[0] != "https://wikipedia.org" {
		t.Fatalf("first link should be https://wikipedia.org not %s", links[0])
	}

	if links[1] != "https://miro.medium.com/v2/resize:fit:900/0*L31Zh4YhAv3Wokco" {
		t.Fatalf("second link should be https://miro.medium.com/v2/resize:fit:900/0*L31Zh4YhAv3Wokco not %s", links[1])
	}

	output := markup.Render(50)
	expected := style.Link("Here's a link!", 1) + "\n\n" +
		style.LinkBlock("This is a beautiful image!", 2) + "\n\n" +
		style.Bullet("Nested list\n"+style.Bullet("Nesting"))

	if expected != output {
		t.Fatalf("expected %s not %s", expected, output)
	}
}
