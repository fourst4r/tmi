package tmi

import (
	"bytes"
	"io"
)

type Prefixer struct {
	prefixFunc      func() string
	writer          io.Writer
	trailingNewline bool
	buf             bytes.Buffer // reuse buffer to save allocations
}

// New creates a new Prefixer that forwards all calls to Write() to writer.Write() with all lines prefixed with the
// return value of prefixFunc. Having a function instead of a static prefix allows to print timestamps or other changing
// information.
func NewPrefixer(writer io.Writer, prefixFunc func() string) *Prefixer {
	return &Prefixer{prefixFunc: prefixFunc, writer: writer, trailingNewline: true}
}

func (pf *Prefixer) Write(payload []byte) (int, error) {
	pf.buf.Reset() // clear the buffer

	for _, b := range payload {
		if pf.trailingNewline {
			pf.buf.WriteString(pf.prefixFunc())
			pf.trailingNewline = false
		}

		pf.buf.WriteByte(b)

		if b == '\n' {
			// do not print the prefix right after the newline character as this might
			// be the very last character of the stream and we want to avoid a trailing prefix.
			pf.trailingNewline = true
		}
	}

	n, err := pf.writer.Write(pf.buf.Bytes())
	if err != nil {
		// never return more than original length to satisfy io.Writer interface
		if n > len(payload) {
			n = len(payload)
		}
		return n, err
	}

	// return original length to satisfy io.Writer interface
	return len(payload), nil
}

// func prefixwriter(w io.Writer) io.Writer {
// 	return &prefixWriter{w}
// }

// type prefixWriter struct {
// 	w io.Writer
// }

// func (p *prefixWriter) Write(a []byte) (n int, err error) {
// 	io.TeeReader()
// 	return
// }
