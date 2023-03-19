package parse

import (
	"bytes"
)

// tw (track writer) tracks how far indented the current write is.
type tw struct {
	width    int
	col      int // current column position
	w        *bytes.Buffer
	bracecol int // track open brace for lists.
}

func (t *tw) Write(p []byte) (int, error) {
	// find last newline in buf and get until end of buf
	last := bytes.LastIndex(p, []byte("\n"))
	if last != -1 {
		t.col = len(p) - last - 1 // TODO(miek), -1 correct here?
	} else {
		t.col += len(p)
	}

	if t.col < 0 {
		panic("tw: col < 0")
	}
	return t.w.Write(p)
}
