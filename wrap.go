package cf

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// wrap reformats comments to fill width. It does not take any special text structures into account.
func wrap(comments []string, indent string, width int) string {
	text := strings.TrimSpace(comments[0][1:]) // remove #
	for i := 1; i < len(comments); i++ {
		text += " " + strings.TrimSpace(comments[i][1:])
	}

	wrap := make([]byte, 0, len(text)+2*len(text)/width)
	eoLine := width
	inWord := false
	for i, j := 0, 0; ; {
		r, size := utf8.DecodeRuneInString(text[i:])
		if size == 0 && r == utf8.RuneError {
			r = ' '
		}
		if unicode.IsSpace(r) {
			if inWord {
				if i >= eoLine {
					wrap = append(wrap, '\n')
					wrap = append(wrap, []byte(indent)...)
					wrap = append(wrap, []byte("# ")...)
					eoLine = len(wrap) + width
				} else if len(wrap) > 0 {
					wrap = append(wrap, ' ')
				}
				wrap = append(wrap, text[j:i]...)
			}
			inWord = false
		} else if !inWord {
			inWord = true
			j = i
		}
		if size == 0 && r == ' ' {
			break
		}
		i += size
	}
	return indent + "# " + string(wrap) // first line
}
