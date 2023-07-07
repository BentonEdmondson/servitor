package ansi

import (
	"fmt"
	"testing"
)

func TestWrap(t *testing.T) {
	// These test were pulled and modified from:
	// https://github.com/muesli/reflow/blob/d4603be2c4a9017b4cf38856841116ffe0f04c59/wordwrap/wordwrap_test.go
	tests := []struct {
		Input    string
		Expected string
		Limit    int
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
			" This\nis a\nlist: \n\n * foo\n * bar\n\n\n * foo\nbar",
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
		// Many, many newlines shouldn't collapse:
		{
			"multi-space   \n\n\n\n\n far down",
			"multi-sp\nace   \n\n\n\n\n far\ndown",
			8,
		},
	}

	for _, test := range tests {
		output := Wrap(test.Input, test.Limit)
		if test.Expected != output {
			t.Fatalf("expected %s but got %s", test.Expected, output)
		}

		// Test that `Wrap` is idempotent
		identical := Wrap(test.Expected, test.Limit)
		if test.Expected != identical {
			t.Fatalf("expected %s but got %s", test.Expected, identical)
		}
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
	if expected != indented {
		t.Fatalf("expected %s but got %s", expected, indented)
	}

	fmt.Println("This should look like a nice, indented code block:")
	styled := Indent(Apply(padded, "48;2;75;75;75"), "  ", true)
	fmt.Println(styled)
}

func TestSnip(t *testing.T) {
	// These test were pulled and modified from:
	// https://github.com/muesli/reflow/blob/d4603be2c4a9017b4cf38856841116ffe0f04c59/wordwrap/wordwrap_test.go
	tests := []struct {
		Input    string
		Expected string
		Height   int
		Width    int
	}{
		// Restrict lines down:
		{
			"one\n\nthree\nfour",
			"one\n\nthree…",
			3,
			25,
		},
		// Don't restrict lines when not necessary:
		{
			"one\n\nthree\nfour",
			"one\n\nthree\nfour",
			5,
			25,
		},
		// Remove last character to insert ellipsis:
		{
			"one\ntwo\nthree\nfour",
			"one\ntwo\nthre…",
			3,
			5,
		},
		// Omit trailing whitespace only lines:
		{
			"one\n\n \nfour",
			"one…",
			3,
			25,
		},
		// Omit trailing whitespace and last character for ellipsis:
		{
			"one\n\n \nfour",
			"on…",
			3,
			3,
		},
		// Omit ellipsis when perfect fit
		{
			"one\ntwo\nthree",
			"one\ntwo\nthree",
			3,
			5,
		},
	}

	for _, test := range tests {
		output := Snip(test.Input, test.Width, test.Height, "…")
		
		if test.Expected != output {
			t.Fatalf("expected %s but got %s", test.Expected, output)
		}
	}
}

func TestCenterVertically(t *testing.T) {
	tests := []struct {
		prefix   string
		centered string
		suffix   string
		height   uint
		output   string
	}{
		// normal case
		{
			"p1\np2",
			"c1\nc2",
			"s1\ns2",
			6,
			"p1\np2\nc1\nc2\ns1\ns2",
		},

		// offset center with even height
		{
			"p1",
			"c1",
			"s1\ns2",
			4,
			"p1\nc1\ns1\ns2",
		},

		// offset center with odd height
		{
			"p1",
			"c1\nc2",
			"s1\ns2",
			5,
			"p1\nc1\nc2\ns1\ns2",
		},

		// trimmed top
		{
			"p1\np2",
			"c1\nc2",
			"s1",
			4,
			"p2\nc1\nc2\ns1",
		},

		// buffered top (with offset)
		{
			"p1",
			"c1",
			"s1\ns2\ns3",
			6,
			"\np1\nc1\ns1\ns2\ns3",
		},

		// trimmed bottom
		{
			"p1",
			"c1",
			"s1\ns2",
			3,
			"p1\nc1\ns1",
		},

		// buffered bottom
		{
			"p1",
			"c1",
			"",
			3,
			"p1\nc1\n",
		},

		// center too big
		{
			"top",
			"middle\nis\nbig",
			"bottom",
			2,
			"middle\nis",
		},

		// perfect center
		{
			"top",
			"middle\nis\nbig",
			"bottom",
			3,
			"middle\nis\nbig",
		},
	}

	for i, test := range tests {
		actual := CenterVertically(test.prefix, test.centered, test.suffix, test.height)
		if test.output != actual {
			t.Fatalf("Expected %v but received %v for test %v", test.output, actual, i)
		}
	}
}
