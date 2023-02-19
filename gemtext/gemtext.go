package gemtext

import (
	"mimicry/style"
	"strings"
	"regexp"
	"errors"
)

/*
	Specification:
	https://gemini.circumlunar.space/docs/specification.html
*/

func Render(text string) (string, error) {
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

		switch {
		case strings.HasPrefix(line, "###"):
			result += style.Header(strings.TrimLeft(line[3:], " \t"), 3) + "\n"
		case strings.HasPrefix(line, "##"):
			result += style.Header(strings.TrimLeft(line[2:], " \t"), 2) + "\n"
		case strings.HasPrefix(line, "#"):
			result += style.Header(strings.TrimLeft(line[1:], " \t"), 1) + "\n"
		case strings.HasPrefix(line, ">"):
			result += style.QuoteBlock(strings.TrimLeft(line[1:], " \t")) + "\n"
		case strings.HasPrefix(line, "* "):
			result += style.Bullet(line[2:]) + "\n"
		case strings.HasPrefix(line, "=>"):
			rendered, err := renderLink(strings.TrimLeft(line[2:], " \t"))
			if err != nil {
				return "", err
			}
			result += rendered + "\n"
		default:
			result += line + "\n"
		}
	}

	// If trailing backticks are omitted, implicitly automatically add them
	if preformattedMode {
		result += style.CodeBlock(strings.TrimSuffix(preformattedBuffer, "\n")) + "\n"
	}

	return strings.Trim(result, " \t\r\n"), nil
}

func renderLink(text string) (string, error) {
	r := regexp.MustCompile(`^(.*?)(?:[ \t]+(.*))?$`)
	matches := r.FindStringSubmatch(text)
	url := matches[1]
	alt := matches[2]

	if alt == "" {
		alt = url
	}

	// another option here is to: return text
	if alt == "" {
		return "", errors.New("Link line with no content found in gemtext")
	}

	return style.LinkBlock(alt), nil
}
