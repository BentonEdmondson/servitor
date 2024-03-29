package style

import (
	"servitor/ansi"
	"strconv"
	"strings"
	"servitor/config"
)

func background(text string, rgb string) string {
	prefix := "48;2;" + rgb
	return ansi.Apply(text, prefix)
}

func foreground(text string, rgb string) string {
	prefix := "38;2;" + rgb
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
	return background(text, config.Parsed.Style.Colors.Code)
}

func Highlight(text string) string {
	return background(text, config.Parsed.Style.Colors.Highlight)
}

func Color(text string) string {
	return foreground(text, config.Parsed.Style.Colors.Primary)
}

func Problem(issue error) string {
	return Red(issue.Error())
}

func Red(text string) string {
	return foreground(text, config.Parsed.Style.Colors.Error)
}

func Link(text string, number int) string {
	return Color(Underline(text) + superscript(number))
}

func CodeBlock(text string) string {
	return Code(text)
}

func QuoteBlock(text string) string {
	prefixed := ansi.Indent(text, "▌", true)
	return Color(prefixed)
}

func LinkBlock(text string, number int) string {
	return "‣ " + ansi.Indent(Link(text, number), "  ", false)
}

func Header(text string, level uint) string {
	indented := ansi.Indent(text, strings.Repeat(" ", int(level+1)), false)
	withPrefix := strings.Repeat("⯁", int(level)) + " " + indented
	return Color(Bold(withPrefix))
}

func Bullet(text string) string {
	return "• " + ansi.Indent(text, "  ", false)
}

func superscript(value int) string {
	text := strconv.Itoa(value)
	return strings.Map(func(input rune) rune {
		switch input {
		case '0':
			return '\u2070'
		case '1':
			return '\u00B9'
		case '2':
			return '\u00B2'
		case '3':
			return '\u00B3'
		case '4':
			return '\u2074'
		case '5':
			return '\u2075'
		case '6':
			return '\u2076'
		case '7':
			return '\u2077'
		case '8':
			return '\u2078'
		case '9':
			return '\u2079'
		default:
			panic("can't superscript non-digit")
		}
	}, text)
}
