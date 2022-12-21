package style

import (
	"fmt"
)

func Display(text string, code int) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", code, text)
}

func Bold(text string) string {
	return Display(text, 1)
}