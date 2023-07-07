package ansi

import (
	"regexp"
	"strings"
	"unicode"
)

func expand(text string) [][]string {
	r := regexp.MustCompile(`(?s)((?:\x1b\[.*?m)*)(.)(?:\x1b\[0m)?`)
	return r.FindAllStringSubmatch(text, -1)
}

func collapse(expanded [][]string) string {
	output := ""
	for _, match := range expanded {
		output += match[0]
	}
	return output
}

func Apply(text string, style string) string {
	expanded := expand(text)
	result := ""
	for _, match := range expanded {
		prefix := match[1]
		letter := match[2]

		if letter == "\n" {
			result += "\n"
			continue
		}

		result += "\x1b[" + style + "m" + prefix + letter + "\x1b[0m"
	}
	return result
}

func Indent(text string, prefix string, includeFirst bool) string {
	expanded := expand(text)
	result := ""

	if includeFirst {
		result = prefix
	}

	for _, match := range expanded {
		full := match[0]
		letter := match[2]

		if letter == "\n" {
			result += "\n" + prefix
			continue
		}

		result += full
	}
	return result
}

const suffix = " "

func Pad(text string, length int) string {
	expanded := expand(text)
	result := ""
	lineLength := 0

	for _, match := range expanded {
		full := match[0]
		letter := match[2]

		if letter == "\n" {
			amount := length - lineLength
			if amount <= 0 {
				result += "\n"
				lineLength = 0
				continue
			}
			result += strings.Repeat(suffix, amount) + "\n"
			lineLength = 0
			continue
		}

		lineLength += 1
		result += full
	}

	/* Final line */
	amount := length - lineLength
	if amount > 0 {
		result += strings.Repeat(suffix, amount)
	}

	return result
}

/*
I am not convinced this works perfectly, but it is well-tested,
so I will call it good for now.
*/
func Wrap(text string, length int) string {
	expanded := expand(text)
	result := []string{}
	var line, space, word string
	var lineLength, spaceLength, wordLength int

	for _, match := range expanded {
		full := match[0]
		letter := match[2]

		if !unicode.IsSpace([]rune(letter)[0]) {
			if wordLength == length {
				/*
					Word fills an entire line; push it as a line
					(we know this won't clobber stuff in `line`, because the word has
					already necessarily forced line to be pushed)
				*/
				result = append(result, word)
				line = ""
				lineLength = 0
				space = ""
				spaceLength = 0
				word = ""
				wordLength = 0
			}

			if lineLength+spaceLength+wordLength >= length {
				/* The word no longer fits on the current line; push the current line */
				result = append(result, line)
				line = ""
				lineLength = 0
				space = ""
				spaceLength = 0
			}

			word += full
			wordLength += 1
			continue
		}

		/* This means whitespace has been encountered; if there's a word, add it to the line */
		if wordLength > 0 {
			line += space + word
			lineLength += spaceLength + wordLength
			space = ""
			spaceLength = 0
			word = ""
			wordLength = 0
		}

		if letter == "\n" {
			/*
				If the spaces can be jammed into the line, add them.
				This ensures that Wrap(Pad(*)) doesn't eliminate the
				padding.
			*/
			if lineLength+spaceLength <= length {
				line += space
				lineLength += spaceLength
			}

			/* Add the current line as-is and clear everything */
			result = append(result, line)
			line = ""
			lineLength = 0
			space = ""
			spaceLength = 0
			word = ""
			wordLength = 0
		} else {
			space += full
			spaceLength += 1
		}
	}

	/* Cleanup */
	if wordLength > 0 {
		line += space + word
		lineLength += spaceLength + wordLength
	}
	finalLetter := ""
	if len(expanded) > 0 {
		finalLetter = expanded[len(expanded)-1][2]
	}
	if lineLength > 0 || finalLetter == "\n" {
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

func DumbWrap(text string, width int) string {
	expanded := expand(text)
	result := ""
	currentLineLength := 0

	for _, match := range expanded {
		full := match[0]
		letter := match[2]

		if letter == "\n" {
			currentLineLength = 0
			result += "\n"
			continue
		}

		if currentLineLength == width {
			currentLineLength = 0
			result += "\n"
		}

		result += full
		currentLineLength += 1
	}
	return result
}

/*
	Limits `text` to the given `height` and `width`, adding an
	ellipsis to the end and omitting trailing whitespace-only lines
*/
func Snip(text string, width, height int, ellipsis string) string {
	snipped := make([]string, 0, height)

	/* This split is fine because newlines are
	   guaranteed to not be wrapped in ansi codes */
	lines := strings.Split(text, "\n")

	requiresEllipsis := false

	if len(lines) <= height {
		height = len(lines)
	} else {
		requiresEllipsis = true
	}

	/* Adding from back to front */
	for i := height - 1; i >= 0; i -= 1 {
		line := expand(lines[i])
		if len(snipped) == 0 {
			if lineIsOnlyWhitespace(line) {
				requiresEllipsis = true
				continue
			}

			/* Remove last character to make way for ellipsis */
			if len(line) == width && requiresEllipsis {
				line = line[:len(line)-1]
			}
		}

		snipped = append([]string{collapse(line)}, snipped...)
	}

	output := strings.Join(snipped, "\n")

	if requiresEllipsis {
		output += ellipsis
	}

	return output
}

func lineIsOnlyWhitespace(expanded [][]string) bool {
	for _, match := range expanded {
		if !unicode.IsSpace([]rune(match[2])[0]) {
			return false
		}
	}

	return true
}

func Height(text string) uint {
	return uint(strings.Count(text, "\n")) + 1
}

func CenterVertically(prefix, centered, suffix string, height uint) string {
	prefixHeight, centeredHeight, suffixHeight := Height(prefix), Height(centered), Height(suffix)
	if height <= centeredHeight {
		return strings.Join(strings.Split(centered, "\n")[:height], "\n")
	}
	totalBufferSize := height - centeredHeight
	topBufferSize := totalBufferSize / 2
	bottomBufferSize := topBufferSize + totalBufferSize%2

	if topBufferSize > prefixHeight {
		prefix = strings.Repeat("\n", int(topBufferSize-prefixHeight)) + prefix
	} else if topBufferSize < prefixHeight {
		prefix = strings.Join(strings.Split(prefix, "\n")[prefixHeight-topBufferSize:], "\n")
	}

	if bottomBufferSize > suffixHeight {
		suffix += strings.Repeat("\n", int(bottomBufferSize-suffixHeight))
	} else if bottomBufferSize < suffixHeight {
		suffix = strings.Join(strings.Split(suffix, "\n")[:bottomBufferSize], "\n")
	}

	return prefix + "\n" + centered + "\n" + suffix
}

func ReplaceLastLine(original, replacement string) string {
	if strings.Contains(replacement, "\n") {
		panic("new version of last line cannot contain a newline")
	}

	var lastIndex = strings.LastIndex(original, "\n")
	if lastIndex == -1 {
		lastIndex = 0
	}
	return original[:lastIndex] + "\n" + replacement
}

func SetLength(text string, length int, ellipsis string) string {
	if length == 0 {
		return ""
	}
	if len(text) > length {
		return text[:length - 1] + ellipsis
	}
	if len(text) < length {
		return text + strings.Repeat(" ", length - len(text))
	}
	return text
}

func Squash(text string) string {
	return strings.ReplaceAll(text, "\n", " ")
}

func Scrub(text string) string {
	text = strings.ReplaceAll(text, "\t", "    ")
	text = strings.Map(func(input rune) rune {
		if input != '\n' && unicode.IsControl(input) {
			return -1
		}
		return input
	}, text)
	return text
}
