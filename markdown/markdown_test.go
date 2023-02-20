package markdown

import (
	"mimicry/util"
	"testing"
	"mimicry/style"
)

func TestBasic(t *testing.T) {
	input := `[Here's a link!](https://wikipedia.org)

![This is a beautiful image!](https://upload.wikimedia.org/wikipedia/commons/thumb/c/cb/Francesco_Melzi_-_Portrait_of_Leonardo.png/800px-Francesco_Melzi_-_Portrait_of_Leonardo.png)

* Nested list
  * Nesting`
	output, err := Render(input)
	if err != nil {
		panic(err)
	}

	expected := style.Link("Here's a link!") + "\n\n" +
		style.LinkBlock("This is a beautiful image!") + "\n\n" +
		style.Bullet("Nested list\n" + style.Bullet("Nesting"))

	util.AssertEqual(expected, output, t)
}