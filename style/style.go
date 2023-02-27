package style

import (
	"fmt"
	"strings"
	"mimicry/ansi"
)

func background(text string, r uint8, g uint8, b uint8) string {
	prefix := fmt.Sprintf("48;2;%d;%d;%d", r, g, b)
	return ansi.Apply(text, prefix)
}

func foreground(text string, r uint8, g uint8, b uint8) string {
	prefix := fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
	return ansi.Apply(text, prefix)
}

func Bold(text string) string {
	return ansi.Apply(text, "1")
}

func Strikethrough(text string) string {
	return ansi.Apply(text, "9")
}

func Underline(text string) string {
	return ansi.Apply(text, "4")
}

func Italic(text string) string {
	return ansi.Apply(text, "3")
}

func Code(text string) string {
	return background(text, 75, 75, 75)
}

func Highlight(text string) string {
	return background(text, 13, 125, 0)
}

func Color(text string) string {
	return foreground(text, 164, 245, 155)
}

func Link(text string) string {
	return Underline(Color(text))
}

func CodeBlock(text string) string {
	return Code(text)
}

func QuoteBlock(text string) string {
	prefixed := ansi.Indent(text, "▌", true)
	return Color(prefixed)
}

func LinkBlock(text string) string {
	indented := ansi.Indent(text, "  ", false)
	return "‣ " + Link(indented)
}

func Header(text string, level uint) string {
	indented := ansi.Indent(text, strings.Repeat(" ", int(level+1)), false)
	withPrefix := strings.Repeat("⯁", int(level)) + " " + indented
	return Color(Bold(withPrefix))
}

func Bullet(text string) string {
	return "• " + ansi.Indent(text, "  ", false)
}
