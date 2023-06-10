package hypertext

import (
	"mimicry/style"
	"mimicry/util"
	"testing"
)

func TestMergeText(t *testing.T) {
	lhs0 := "front"
	rhs0 := "back"
	output0 := mergeText(lhs0, rhs0)
	expected0 := "frontback"
	util.AssertEqual(expected0, output0, t)

	lhs1 := "front     "
	rhs1 := "   back"
	output1 := mergeText(lhs1, rhs1)
	expected1 := "front back"
	util.AssertEqual(expected1, output1, t)

	lhs2 := "front     "
	rhs2 := " \n  back"
	output2 := mergeText(lhs2, rhs2)
	expected2 := "front\nback"
	util.AssertEqual(expected2, output2, t)

	lhs3 := "front    \n\n\n "
	rhs3 := " \n  back"
	output3 := mergeText(lhs3, rhs3)
	expected3 := "front\n\nback"
	util.AssertEqual(expected3, output3, t)
}

func TestStyles(t *testing.T) {
	input := "<s>s</s><code>code</code><i>i</i><u>u</u><mark>mark</mark>"
	output, err := Render(input, 50)
	if err != nil {
		panic(err)
	}
	expected := style.Strikethrough("s") +
		style.Code("code") +
		style.Italic("i") +
		style.Underline("u") +
		style.Highlight("mark")

	util.AssertEqual(expected, output, t)
}

func TestSurroundingBlocks(t *testing.T) {
	input := "<p>first</p>in \t<mark>the</mark> \rmiddle<p>last</p>"
	output, err := Render(input, 50)
	if err != nil {
		panic(err)
	}
	expected := `first

in ` + style.Highlight("the") + ` middle

last`
	util.AssertEqual(expected, output, t)
}

func TestAdjacentBlocks(t *testing.T) {
	input := "\t<p>first</p>\n\t<p>second</p>"
	output, err := Render(input, 50)
	if err != nil {
		panic(err)
	}
	expected := `first

second`
	util.AssertEqual(expected, output, t)
}

func TestPoetry(t *testing.T) {
	input := "he shouted\t\ta few words<br>at those annoying birds<br><br>and that they heard"
	output, err := Render(input, 50)
	if err != nil {
		panic(err)
	}
	expected := `he shouted a few words
at those annoying birds

and that they heard`

	util.AssertEqual(expected, output, t)
}

// TODO: this is broken for now because my wrap algorithm removes
// trailing spaces under certain conditions. I need to modify it such that it
// leaves trailing spaces if possible
func TestPreservation(t *testing.T) {
	input := "<pre>multi-space   \n\n\n\n\n far down</pre>"
	output, err := Render(input, 50)
	if err != nil {
		panic(err)
	}
	expected := style.CodeBlock(`multi-space   




 far down`)
	util.AssertEqual(expected, output, t)
}

func TestNestedBlocks(t *testing.T) {
	input := `<p>Once a timid child</p>

<p> </p>

<p><img src="https://i.snap.as/P8qpdMbM.jpg" alt=""/></p>`
	output, err := Render(input, 50)
	if err != nil {
		panic(err)
	}
	expected := `Once a timid child

` + style.LinkBlock("https://i.snap.as/P8qpdMbM.jpg")
	util.AssertEqual(expected, output, t)
}

func TestAdjacentLists(t *testing.T) {
	input := `<ul><li>top list</li></ul><ul><li>bottom list</li></ul>`
	output, err := Render(input, 50)
	if err != nil {
		panic(err)
	}
	expected := style.Bullet("top list") + "\n\n" +
		style.Bullet("bottom list")
	util.AssertEqual(expected, output, t)
}

func TestNestedLists(t *testing.T) {
	input := `<ul><li>top list<ul><li>nested</li></ul></li></ul>`
	output, err := Render(input, 50)
	if err != nil {
		panic(err)
	}
	expected := style.Bullet("top list\n" + style.Bullet("nested"))

	util.AssertEqual(expected, output, t)
}

func TestBlockInList(t *testing.T) {
	input := `<ul><li>top list<p><ul><li>paragraph</li></ul></p></li></ul>`
	output, err := Render(input, 50)
	if err != nil {
		panic(err)
	}
	expected := style.Bullet("top list\n\n" + style.Bullet("paragraph"))

	util.AssertEqual(expected, output, t)
}
