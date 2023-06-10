package gemtext

import (
	"mimicry/style"
	"regexp"
	"strings"
)

/*
	Specification:
	https://gemini.circumlunar.space/docs/specification.html
*/

func Render(text string, width int) (string, error) {
	lines := strings.Split(text, "\n")
	result := ""
	preformattedMode := false
	preformattedBuffer := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if preformattedMode {
				result += style.CodeBlock(strings.TrimSuffix(preformattedBuffer, "\n")) + "\n"
				preformattedBuffer = ""
				preformattedMode = false
			} else {
				preformattedMode = true
			}
			continue
		}

		if preformattedMode {
			preformattedBuffer += line + "\n"
			continue
		}

		if match := regexp.MustCompile(`^=>[ \t]*(.*?)(?:[ \t]+(.*))?$`).FindStringSubmatch(line); len(match) == 3 {
			url := match[1]
			alt := match[2]
			if alt == "" {
				alt = url
			}
			result += style.LinkBlock(alt) + "\n"
		} else if match := regexp.MustCompile(`^#[ \t]+(.*)$`).FindStringSubmatch(line); len(match) == 2 {
			result += style.Header(match[1], 1) + "\n"
		} else if match := regexp.MustCompile(`^##[ \t]+(.*)$`).FindStringSubmatch(line); len(match) == 2 {
			result += style.Header(match[1], 2) + "\n"
		} else if match := regexp.MustCompile(`^###[ \t]+(.*)$`).FindStringSubmatch(line); len(match) == 2 {
			result += style.Header(match[1], 3) + "\n"
		} else if match := regexp.MustCompile(`^\* (.*)$`).FindStringSubmatch(line); len(match) == 2 {
			result += style.Bullet(match[1]) + "\n"
		} else if match := regexp.MustCompile(`^> ?(.*)$`).FindStringSubmatch(line); len(match) == 2 {
			result += style.QuoteBlock(match[1]) + "\n"
		} else {
			result += line + "\n"
		}
	}

	// If trailing backticks are omitted, implicitly automatically add them
	if preformattedMode {
		result += style.CodeBlock(strings.TrimSuffix(preformattedBuffer, "\n")) + "\n"
	}

	return strings.TrimSuffix(result, "\n"), nil
}
