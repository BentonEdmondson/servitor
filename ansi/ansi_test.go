package ansi

import (
	"testing"
	"mimicry/util"
	"fmt"
)

func TestBasic(t *testing.T) {
	// These test were pulled and modified from:
	// https://github.com/muesli/reflow/blob/d4603be2c4a9017b4cf38856841116ffe0f04c59/wordwrap/wordwrap_test.go
	tests := []struct {
		Input        string
		Expected     string
		Limit        int
	}{
		// Nothing to wrap here, should pass through:
		{
			"foo",
			"foo",
			4,
		},
		// Snap words
		{
			"foobarfoo",
			"foob\narfo\no",
			4,
		},
		// Lines are broken at whitespace:
		{
			"foo bar foo",
			"foo\nbar\nfoo",
			4,
		},
		// Space buffer needs to be emptied before breakpoints:
		{
			"foo --bar",
			"foo --bar",
			9,
		},
		// Lines are broken at whitespace, and long words break as well
		{
			"foo bars foobars",
			"foo\nbars\nfoob\nars",
			4,
		},
		// A word that would run beyond the limit is wrapped:
		{
			"foo bar",
			"foo\nbar",
			5,
		},
		// Whitespace prefixing an explicit line break remains:
		{
			"foo\nb  a\n bar",
			"foo\nb  a\n bar",
			4,
		},
		// Trailing whitespace is removed if it doesn't fit the width.
		// Runs of whitespace on which a line is broken are removed:
		{
			"foo    \nb   ar   ",
			"foo\nb\nar",
			4,
		},
		// An explicit line break at the end of the input is preserved:
		{
			"foo bar foo\n",
			"foo\nbar\nfoo\n",
			4,
		},
		// Explicit break are always preserved:
		{
			"\nfoo bar\n\n\nfoo\n",
			"\nfoo\nbar\n\n\nfoo\n",
			4,
		},
		// Complete example:
		{
			" This is a list: \n\n * foo\n * bar\n\n\n * foo  \nbar    ",
			" This\nis a\nlist:\n\n * foo\n * bar\n\n\n * foo\nbar",
			6,
		},
		// ANSI sequence codes don't affect length calculation:
		{
			"\x1B[38;2;249;38;114mfoo\x1B[0m\x1B[38;2;248;248;242m \x1B[0m\x1B[38;2;230;219;116mbar\x1B[0m",
			"\x1B[38;2;249;38;114mfoo\x1B[0m\x1B[38;2;248;248;242m \x1B[0m\x1B[38;2;230;219;116mbar\x1B[0m",
			7,
		},
		// ANSI control codes don't get wrapped:
		{
			"\x1B[38;2;249;38;114m(\x1B[0m\x1B[38;2;248;248;242mjust another test\x1B[38;2;249;38;114m)\x1B[0m",
			"\x1B[38;2;249;38;114m(\x1B[0m\x1B[38;2;248;248;242mju\nst\nano\nthe\nr\ntes\nt\x1B[38;2;249;38;114m)\x1B[0m",
			3,
		},
	}

	for _, test := range tests {
		output := Wrap(test.Input, test.Limit)
		util.AssertEqual(test.Expected, output, t)

		// Test that `Wrap` is idempotent
		identical := Wrap(test.Expected, test.Limit)
		util.AssertEqual(test.Expected, identical, t)
	}
}

func TestCodeBlock(t *testing.T) {
	input := "Soft-wrapped code block used to test everything"
	wrapped := Wrap(input, 6)
	padded := Pad(wrapped, 6)
	indented := Indent(padded, "  ", true)
	expected := `  Soft-w
  rapped
  code  
  block 
  used  
  to    
  test  
  everyt
  hing  `
	util.AssertEqual(expected, indented, t)

	fmt.Println("This should look like a nice, indented code block:")
	styled := Indent(Apply(padded, "48;2;75;75;75"), "  ", true)
	fmt.Println(styled)
}
