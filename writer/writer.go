package writer

import (
	"io"

	"github.com/sysdevguru/bufnet/limiter"
)

// Writer is a wrapper for io.Writer
type Writer struct {
	lim limiter.Limiter
	dst io.Writer
}

// NewWriter returns writer with bandwidth limited
func NewWriter(d io.Writer, bandwidth int) *Writer {
	w := &Writer{
		dst: d,
		lim: limiter.Limiter{Bandwidth: bandwidth},
	}
	return w
}

// Write implements the io.Writer and maintains the given bandwidth.
func (w *Writer) Write(p []byte) (n int, err error) {
	w.lim.Init()

	n, err = w.dst.Write(p)
	if err != nil {
		return n, err
	}

	w.lim.Limit(n, len(p))

	return n, err
}
