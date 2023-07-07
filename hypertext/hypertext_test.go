package hypertext

import (
	"mimicry/ansi"
	"mimicry/style"
	"testing"
)

func TestMergeText(t *testing.T) {
	lhs0 := "front"
	rhs0 := "back"
	output0 := mergeText(lhs0, rhs0)
	expected0 := "frontback"
	if expected0 != output0 {
		t.Fatalf("expected %s not %s", expected0, output0)
	}

	lhs1 := "front     "
	rhs1 := "   back"
	output1 := mergeText(lhs1, rhs1)
	expected1 := "front back"
	if expected1 != output1 {
		t.Fatalf("expected %s not %s", expected1, output1)
	}

	lhs2 := "front     "
	rhs2 := " \n  back"
	output2 := mergeText(lhs2, rhs2)
	expected2 := "front\nback"
	if expected2 != output2 {
		t.Fatalf("expected %s not %s", expected2, output2)
	}

	lhs3 := "front    \n\n\n "
	rhs3 := " \n  back"
	output3 := mergeText(lhs3, rhs3)
	expected3 := "front\n\nback"
	if expected3 != output3 {
		t.Fatalf("expected %s not %s", expected3, output3)
	}
}

func TestStyles(t *testing.T) {
	input := "<s>s</s><code>code</code><i>i</i><u>u</u><mark>mark</mark>"
	markup, _, err := NewMarkup(input)
	if err != nil {
		t.Fatal(err)
	}
	output := markup.Render(50)
	if err != nil {
		t.Fatal(err)
	}
	expected := style.Strikethrough("s") +
		style.Code("code") +
		style.Italic("i") +
		style.Underline("u") +
		style.Highlight("mark")

	if expected != output {
		t.Fatalf("excpected output to be %s not %s", expected, output)
	}
}

func TestSurroundingBlocks(t *testing.T) {
	input := "<p>first</p>in \t<mark>the</mark> \rmiddle<p>last</p>"
	markup, _, err := NewMarkup(input)
	if err != nil {
		t.Fatal(err)
	}
	output := markup.Render(50)
	if err != nil {
		t.Fatal(err)
	}
	expected := `first

in ` + style.Highlight("the") + ` middle

last`
	if expected != output {
		t.Fatalf("excpected output to be %s not %s", expected, output)
	}
}

func TestAdjacentBlocks(t *testing.T) {
	input := "\t<p>first</p>\n\t<p>second</p>"
	markup, _, err := NewMarkup(input)
	if err != nil {
		t.Fatal(err)
	}
	output := markup.Render(50)
	if err != nil {
		t.Fatal(err)
	}
	expected := `first

second`
	if expected != output {
		t.Fatalf("excpected output to be %s not %s", expected, output)
	}
}

func TestPoetry(t *testing.T) {
	input := "he shouted\t\ta few words<br>at those annoying birds<br><br>and that they heard"
	markup, _, err := NewMarkup(input)
	if err != nil {
		t.Fatal(err)
	}
	output := markup.Render(50)
	if err != nil {
		t.Fatal(err)
	}
	expected := `he shouted a few words
at those annoying birds

and that they heard`

	if expected != output {
		t.Fatalf("excpected output to be %s not %s", expected, output)
	}
}

func TestPreservation(t *testing.T) {
	input := "<pre>multi-space   \n\n\n\n\n far down</pre>"
	markup, _, err := NewMarkup(input)
	if err != nil {
		t.Fatal(err)
	}
	output := markup.Render(50)
	if err != nil {
		t.Fatal(err)
	}
	expected := style.CodeBlock(ansi.Pad(`multi-space   




 far down`, 50))
	if expected != output {
		t.Fatalf("excpected output to be %s not %s", expected, output)
	}
}

func TestNestedBlocks(t *testing.T) {
	input := `<p>Once a timid child</p>

<p> </p>

<p><img src="https://i.snap.as/P8qpdMbM.jpg" alt=""/></p>`
	markup, _, err := NewMarkup(input)
	if err != nil {
		t.Fatal(err)
	}
	output := markup.Render(50)
	if err != nil {
		t.Fatal(err)
	}
	expected := `Once a timid child

` + style.LinkBlock("https://i.snap.as/P8qpdMbM.jpg", 1)
	if expected != output {
		t.Fatalf("excpected output to be %s not %s", expected, output)
	}
}

func TestAdjacentLists(t *testing.T) {
	input := `<ul><li>top list</li></ul><ul><li>bottom list</li></ul>`
	markup, _, err := NewMarkup(input)
	if err != nil {
		t.Fatal(err)
	}
	output := markup.Render(50)
	if err != nil {
		t.Fatal(err)
	}
	expected := style.Bullet("top list") + "\n\n" +
		style.Bullet("bottom list")
	if expected != output {
		t.Fatalf("excpected output to be %s not %s", expected, output)
	}
}

func TestNestedLists(t *testing.T) {
	input := `<ul><li>top list<ul><li>nested</li></ul></li></ul>`
	markup, _, err := NewMarkup(input)
	if err != nil {
		t.Fatal(err)
	}
	output := markup.Render(50)
	if err != nil {
		t.Fatal(err)
	}
	expected := style.Bullet("top list\n" + style.Bullet("nested"))

	if expected != output {
		t.Fatalf("excpected output to be %s not %s", expected, output)
	}
}

func TestBlockInList(t *testing.T) {
	input := `<ul><li>top list<p><ul><li>paragraph</li></ul></p></li></ul>`
	markup, _, err := NewMarkup(input)
	if err != nil {
		t.Fatal(err)
	}
	output := markup.Render(50)
	if err != nil {
		t.Fatal(err)
	}
	expected := style.Bullet("top list\n\n" + style.Bullet("paragraph"))

	if expected != output {
		t.Fatalf("excpected output to be %s not %s", expected, output)
	}
}

func TestWrapping(t *testing.T) {
	input := `<p>hello sir</p>`
	markup, _, err := NewMarkup(input)
	if err != nil {
		t.Fatal(err)
	}
	output := markup.Render(4)
	if err != nil {
		t.Fatal(err)
	}
	expected := "hell\no\nsir"

	if expected != output {
		t.Fatalf("excpected output to be %s not %s", expected, output)
	}
}

func TestLinks(t *testing.T) {
	input := `<a href="https://wikipedia.org">Great site</a>
<img src="https://example.org" alt="What the heck">
<iframe title="Music" src="https://spotify.com">`
	markup, links, err := NewMarkup(input)
	if err != nil {
		t.Fatal(err)
	}

	if links[0] != "https://wikipedia.org" {
		t.Fatalf("the first links should have been https://wikipedia.org not %s", links[0])
	}

	if links[1] != "https://example.org" {
		t.Fatalf("the first links should have been https://example.org not %s", links[1])
	}

	if links[2] != "https://spotify.com" {
		t.Fatalf("the first links should have been https://spotify.com not %s", links[2])
	}

	output := markup.Render(50)
	if err != nil {
		t.Fatal(err)
	}
	expected := style.Link("Great site", 1) + "\n\n" +
		style.LinkBlock("What the heck", 2) + "\n\n" +
		style.LinkBlock("Music", 3)

	if expected != output {
		t.Fatalf("excpected output to be %s not %s", expected, output)
	}
}
