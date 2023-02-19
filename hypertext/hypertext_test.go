package hypertext

import (
	"testing"
	"mimicry/style"
	"mimicry/utils"
)

func TestMergeText(t *testing.T) {
	lhs0 := "front"
	rhs0 := "back"
	output0 := mergeText(lhs0, rhs0)
	expected0 := "frontback"
	utils.AssertEqual(expected0, output0, t)

	lhs1 := "front     "
	rhs1 := "   back"
	output1 := mergeText(lhs1, rhs1)
	expected1 := "front back"
	utils.AssertEqual(expected1, output1, t)

	lhs2 := "front     "
	rhs2 := " \n  back"
	output2 := mergeText(lhs2, rhs2)
	expected2 := "front\nback"
	utils.AssertEqual(expected2, output2, t)

	lhs3 := "front    \n\n\n "
	rhs3 := " \n  back"
	output3 := mergeText(lhs3, rhs3)
	expected3 := "front\n\nback"
	utils.AssertEqual(expected3, output3, t)
}

func TestStyles(t *testing.T) {
	input := "<s>s</s><code>code</code><i>i</i><u>u</u><mark>mark</mark>"
	output, err := Render(input)
	if err != nil {
		panic(err)
	}
	expected := style.Strikethrough("s") +
		style.Code("code") +
		style.Italic("i") +
		style.Underline("u") +
		style.Highlight("mark")

	utils.AssertEqual(expected, output, t)
}

func TestSurroundingBlocks(t *testing.T) {
	input := "<p>first</p>in \t<mark>the</mark> \rmiddle<p>last</p>"
	output, err := Render(input)
	if err != nil {
		panic(err)
	}
	expected := `first

in ` + style.Highlight("the") + ` middle

last`
	utils.AssertEqual(expected, output, t)
}

func TestAdjacentBlocks(t *testing.T) {
	input := "\t<p>first</p>\n\t<p>second</p>"
	output, err := Render(input)
	if err != nil {
		panic(err)
	}
	expected := `first

second`
	utils.AssertEqual(expected, output, t)
}

func TestPoetry(t *testing.T) {
	input := "he shouted\t\ta few words<br>at those annoying birds<br><br>and that they heard"
	output, err := Render(input)
	if err != nil {
		panic(err)
	}
	expected := `he shouted a few words
at those annoying birds

and that they heard`

	utils.AssertEqual(expected, output, t)
}

func TestPreservation(t *testing.T) {
	input := "<pre>tab\tand multi-space   \n\n\n\n\n far down</pre>"
	output, err := Render(input)
	if err != nil {
		panic(err)
	}
	expected := style.CodeBlock(`tab	and multi-space   




 far down`)
	utils.AssertEqual(expected, output, t)
}

func TestNestedBlocks(t *testing.T) {
	input := `<p>Once a timid child</p>

<p> </p>

<p><img src="https://i.snap.as/P8qpdMbM.jpg" alt=""/></p>`
	output, err := Render(input)	
	if err != nil {
		panic(err)
	}
	expected := `Once a timid child

` + style.LinkBlock("https://i.snap.as/P8qpdMbM.jpg")
	utils.AssertEqual(expected, output, t)
}
