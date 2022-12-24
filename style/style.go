package style

import (
	"fmt"
)

/*
	To, e.g., bold, prepend the bold character,
	then substitute all resets with `${reset}${bold}`
	to force rebold after all resets, to make sure. Might
	be complex with layering
*/

// const (
// 	Bold = 
// )

func Display(text string, code int) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", code, text)
}

func Bold(text string) string {
	return Display(text, 1)
}

// func Underline(text string) string {
// 	return Display(text, )
// }

// func Anchor(text string) string {

// }