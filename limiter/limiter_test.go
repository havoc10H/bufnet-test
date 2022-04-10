package limiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type test struct {
	written    int
	bufferSize int
	lim        Limiter
}

func TestLimit(t *testing.T) {
	// wrote 1024 with 1024 bandwidth, wait 1s
	testObj := test{written: 1024, bufferSize: 4096, lim: Limiter{Bandwidth: 1024}}
	lim := Limiter{Bandwidth: testObj.lim.Bandwidth}
	start := time.Now()
	lim.Init()
	lim.Limit(testObj.written, testObj.bufferSize)
	end := time.Now()
	diff := end.Sub(start)
	assert.Equal(t, 1, int(diff.Seconds()))

	// bandwidth is negative, don't need to wait
	testObj = test{written: 1024, bufferSize: 4096, lim: Limiter{Bandwidth: -1024}}
	lim = Limiter{Bandwidth: testObj.lim.Bandwidth}
	start = time.Now()
	lim.Init()
	lim.Limit(testObj.written, testObj.bufferSize)
	end = time.Now()
	diff = end.Sub(start)
	assert.Equal(t, 0, int(diff.Seconds()))

	// wrote 2048 with 1024 bandwidth, wait 2s
	testObj = test{written: 2048, bufferSize: 20, lim: Limiter{Bandwidth: 1024}}
	lim = Limiter{Bandwidth: testObj.lim.Bandwidth}
	start = time.Now()
	lim.Init()
	lim.Limit(testObj.written, testObj.bufferSize)
	end = time.Now()
	diff = end.Sub(start)
	assert.Equal(t, 2, int(diff.Seconds()))

	// 1 byte writing with 1024 bandwidth
	// elapsed time would be 976 Microseconds
	// but it is around 2000 Microseconds
	testObj = test{written: 1, bufferSize: 20, lim: Limiter{Bandwidth: 1024}}
	lim = Limiter{Bandwidth: testObj.lim.Bandwidth}
	start = time.Now()
	lim.Init()
	lim.Limit(testObj.written, testObj.bufferSize)
	end = time.Now()
	diff = end.Sub(start)
	assert.Equal(t, 976, int(diff.Microseconds()))

	// actual elapsed time inside the limiter would be 976 Microseconds
	// 1 / 1024 * 1,000,000
	testObj = test{written: 1, bufferSize: 20, lim: Limiter{Bandwidth: 1024}}
	lim = Limiter{Bandwidth: testObj.lim.Bandwidth}
	start = time.Now()
	lim.Init()
	lim.Limit(testObj.written, testObj.bufferSize)
	end = time.Now()
	diff = end.Sub(start)
	assert.Equal(t, 976, int(lim.KeepTime.Microseconds()))
}
