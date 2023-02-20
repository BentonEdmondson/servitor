package gemtext

import (
	"mimicry/style"
	"strings"
	"regexp"
)

/*
	Specification:
	https://gemini.circumlunar.space/docs/specification.html
*/

// TODO: be careful about collapsing whitespace,
// e.g. block quotes should probably be able to start
// with tabs

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
			/*
				Don't just TrimLeft all whitespace, because indents should be possible,
				but at the same time, most people use "> " before their text, so trim
				a single space if it is present. This is not in the spec but is used
				in Amfora and presumably others:
				https://github.com/makeworld-the-better-one/amfora/blob/0b3f874ef19f652fc587ffa80aa6fd08103f892c/renderer/renderer.go#L295

				This could be annoying if someone writes
					>first line
					>  second line indented
				instead of
					> first line
					>   second line indented
			*/
			result += style.QuoteBlock(strings.TrimPrefix(line[1:], " ")) + "\n"
		case strings.HasPrefix(line, "* "):
			/*
				The spec says nothing about optional whitespace, so don't trim at all.
			*/
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

	return strings.TrimSpace(result), nil
}

func renderLink(text string) (string, error) {
	/*
		Regexp to split the line into the url and the optional
		alt text, while also removing the optional whitespace
	*/
	r := regexp.MustCompile(`^(.*?)(?:[ \t]+(.*))?$`)
	matches := r.FindStringSubmatch(text)
	url := matches[1]
	alt := matches[2]

	if alt == "" {
		alt = url
	}

	if alt == "" {
		return text, nil
	}

	return style.LinkBlock(alt), nil
}
