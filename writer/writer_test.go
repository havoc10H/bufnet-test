package writer

import (
	"io"
	"io/ioutil"
	"testing"
	"time"
)

func TestWrite(t *testing.T) {
	t.Parallel()

	// test writing 1024 * 1024 data with 500 * 1024 buffer
	// expected time is 1.95s ~ 2.15s
	tr := &testReader{Size: 1 << 20}
	bw := NewWriter(ioutil.Discard, 500<<10)

	start := time.Now()
	n, err := io.Copy(bw, tr)
	dur := time.Since(start)
	if err != nil {
		t.Error(err)
	}
	if n != 1<<20 {
		t.Errorf("Want %d bytes, got %d.", 1<<20, n)
	}
	t.Logf("Wrote %d bytes in %s.", n, dur)
	if dur < 1950*time.Millisecond || dur > 2150*time.Millisecond {
		t.Errorf("Took %s, want 1.95s~2.15s.", dur)
	}
}

type testReader struct {
	Size int
}

// Read implements io.Reader interface
func (r *testReader) Read(p []byte) (n int, err error) {
	l := len(p)
	if l < r.Size {
		n = l
	} else {
		n = r.Size
		err = io.EOF
	}
	r.Size -= n
	return n, err
}
