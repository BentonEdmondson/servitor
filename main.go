package main

import (
	"encoding/json"
	"mimicry/kinds"
	"os"
	"fmt"
	// "mimicry/style"
	// "mimicry/render"
)

// TODO: even if only supported in few terminals,
// consider using the proportional spacing codes when possible

// TODO: when returning errors, use zero value for return
// also change all error messages to using sprintf-style
// formatting, all lowercase, and no punctuation

func main() {
	// fmt.Println(style.Bold("Bold") + "\tNot Bold")
	// fmt.Println(style.Strikethrough("Strikethrough") + "\tNot Strikethrough")
	// fmt.Println(style.Underline("Underline") + "\tNot Underline")
	// fmt.Println(style.Italic("Italic") + "\tNot Italic")
	// fmt.Println(style.Code("Code") + "\tNot Code")
	// fmt.Println(style.Highlight("Highlight") + "\tNot Highlight")

	// fmt.Println(style.Highlight("Stuff here " + style.Code("CODE") + " more here"))
	// fmt.Println(style.Bold("struff " + style.Strikethrough("bad") + " more stuff"))

	// fmt.Println(style.Linkify("Hello!"))

	// output, err := render.Render("<p>Hello<code>hi</code> Everyone</p><i>@everyone</i> <blockquote>please<br>don't!</blockquote>", "text/html")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(output)

	link := os.Args[len(os.Args)-1]
	command := os.Args[1]

	content, err := kinds.FetchUnknown(link)
	if err != nil {
		panic(err)
	}

	if command == "raw" {
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(content); err != nil {
			panic(err)
		}
		return
	}

	if str, err := content.String(); err != nil {
		panic(err)
	} else {
		fmt.Println(str)
	}
}
