package ansi

import (
	"regexp"
	"strings"
	"unicode"
)

func expand(text string) [][]string {
	r := regexp.MustCompile(`(?s)(?:(?:\x1b\[.*?m)*)(.)(?:\x1b\[0m)?`)
	return r.FindAllStringSubmatch(text, -1)
}

func Apply(text string, style string) string {
	expanded := expand(text)
	result := ""
	for _, match := range expanded {
		full := match[0]
		letter := match[1]

		if letter == "\n" {
			result += "\n"
			continue
		}

		result += "\x1b[" + style + "m" + full + "\x1b[0m"
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
		letter := match[1]

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
		letter := match[1]

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
		letter := match[1]

		/* TODO: I need to find the list of non-breaking whitespace characters
			to exclude from this conditional */
		if !unicode.IsSpace([]rune(letter)[0]) {
			if wordLength == length {
				/*
					Word fills an entire line; push it as a line
					(we know this won't clobber stuff in `line`, because the word has
					already necessarily forced line to be pushed)
				*/
				result = append(result, word)
				line = ""; lineLength = 0
				space = ""; spaceLength = 0
				word = ""; wordLength = 0
			}
			
			if lineLength + spaceLength + wordLength >= length {
				/* The word no longer fits on the current line; push the current line */
				result = append(result, line)
				line = ""; lineLength = 0
				space = ""; spaceLength = 0
			}

			word += full; wordLength += 1
			continue
		}

		/* This means whitespace has been encountered; if there's a word, add it to the line */
		if wordLength > 0 {
			line += space + word; lineLength += spaceLength + wordLength
			space = ""; spaceLength = 0
			word = ""; wordLength = 0
		}

		if letter == "\n" {
			/* Add the current line as-is and clear everything */
			result = append(result, line)
			line = ""; lineLength = 0
			space = ""; spaceLength = 0
			word = ""; wordLength = 0
		} else {
			space += full; spaceLength += 1
		}
	}

	/* Cleanup */
	if wordLength > 0 {
		line += space + word; lineLength += spaceLength + wordLength
	}
	finalLetter := expanded[len(expanded)-1][1]
	if lineLength > 0 || finalLetter == "\n" {
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

/*
	TODO:
		add `Scrub` function that removes all ANSI codes from text
		(this will be used when people redirect output to file)

		add `Snip` function that limits text to a certain number
		of lines, adding an ellipsis if this required removing
		some text

		add `Squash` function that converts newlines to spaces
		(this will be used to prevent newlines from appearing
		in things like names and titles)
*/
