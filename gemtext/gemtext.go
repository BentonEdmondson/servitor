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

type Markup struct {
	tree []string
	cached string
	cachedWidth int
}

func NewMarkup(text string) (*Markup, []string, error) {
	lines := strings.Split(text, "\n")
	rendered, links := renderWithLinks(lines, 80)
	return &Markup{
		tree: lines,
		cached: rendered,
		cachedWidth: 80,
	}, links, nil
}

func (m *Markup) Render(width int) string {
	if m.cachedWidth == width {
		return m.cached
	}
	rendered, _ := renderWithLinks(m.tree, width)
	m.cached = rendered
	m.cachedWidth = width
	return rendered
}

func renderWithLinks(lines []string, width int) (string, []string) {
	links := []string{}
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
			uri := match[1]
			alt := match[2]
			if alt == "" {
				alt = uri
			}
			links = append(links, uri)
			result += style.LinkBlock(alt, len(links)) + "\n"
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

	return strings.Trim(result, "\n"), links
}
