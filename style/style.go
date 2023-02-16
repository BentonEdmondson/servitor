package style

import (
	"fmt"
	"strings"
)

// TODO: at some point I need to sanitize preexisting escape codes
// in input, to do so replace the escape character with visual
// escape character

/*
	To, e.g., bold, prepend the bold character,
	then substitute all resets with `${reset}${bold}`
	to force rebold after all resets, to make sure. Might
	be complex with layering
*/

// const (
// 	Bold = 
// )

func Background(text string, r uint8, g uint8, b uint8) string {
	setter := fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
	resetter := "\x1b[49m"
	text = strings.ReplaceAll(text, resetter, setter)
	return fmt.Sprintf("%s%s%s", setter, text, resetter)
}

func ExtendBackground(text string) string {
	return strings.ReplaceAll(text, "\n", "\x1b[K\n")
}

func Foreground(text string, r uint8, g uint8, b uint8) string {
	setter := fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
	resetter := "\x1b[39m"
	newText := strings.ReplaceAll(text, resetter, setter)
	return fmt.Sprintf("%s%s%s", setter, newText, resetter)
}

func Display(text string, prependCode int, appendCode int) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[%dm", prependCode, text, appendCode)
}

// 21 doesn't work (does double underline)
// 22 removes bold and faint, faint is never used
// so it does the job
func Bold(text string) string {
	return Display(text, 1, 22)
}

func Strikethrough(text string) string {
	return Display(text, 9, 29)
}

func Underline(text string) string {
	return Display(text, 4, 24)
}

func Italic(text string) string {
	return Display(text, 3, 23)
}

func Code(text string) string {
	return Background(text, 75, 75, 75)
}

func CodeBlock(text string) string {
	return ExtendBackground(Code(text))
}

func Highlight(text string) string {
	return Background(text, 13, 125, 0)
}

func Color(text string) string {
	return Foreground(text, 164, 245, 155)
}

func Linkify(text string) string {
	return Underline(Color(text))
}

// func Underline(text string) string {
// 	return Display(text, )
// }

// func Anchor(text string) string {

// }