package style

import (
	"fmt"
	"strings"
	"mimicry/ansi"
)

// TODO: at some point I need to sanitize preexisting escape codes
// in input, to do so replace the escape character with visual
// escape character

func background(text string, r uint8, g uint8, b uint8) string {
	prefix := fmt.Sprintf("48;2;%d;%d;%d", r, g, b)
	return ansi.Apply(text, prefix)
}

func foreground(text string, r uint8, g uint8, b uint8) string {
	prefix := fmt.Sprintf("38;2;%d;%d;%d", r, g, b)
	return ansi.Apply(text, prefix)
}

func display(text string, prependCode int, appendCode int) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[%dm", prependCode, text, appendCode)
}

// 21 doesn't work (does double underline)
// 22 removes bold and faint, faint is never used
// so it does the job
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

func CodeBlock(text string) string {
	return Code(text)
}

func QuoteBlock(text string) string {
	withBar := "▌" + strings.ReplaceAll(text, "\n", "\n▌")
	return Color(withBar)
}

func LinkBlock(text string) string {
	return "‣ " + Link(text)
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

func Header(text string, level uint) string {
	withPrefix := strings.Repeat("⯁", int(level)) + " " + text
	return Color(Bold(withPrefix))
}

func Bullet(text string) string {
	return "• " + strings.ReplaceAll(text, "\n", "\n  ")
}
