package limiter

import "time"

// Limiter limits bandwidth
type Limiter struct {
	Bandwidth   int
	Bucket      int64
	Initialized bool
	Start       time.Time
	KeepTime    time.Duration
}

// Init initialize Limiter
func (l *Limiter) Init() {
	if !l.Initialized {
		l.reset()
		l.Initialized = true
	}
}

func (l *Limiter) reset() {
	l.Bucket = 0
	l.Start = time.Now()
}

// Limit is the function that actually limits bandwidth
func (l *Limiter) Limit(n, bufSize int) {
	// not apply limit in case desired bandwidth is 0 or negative
	if l.Bandwidth <= 0 {
		return
	}

	l.Bucket += int64(n)

	// elapsed time for the read/write operation
	elapsed := time.Since(l.Start)
	// sleep for the keeped time and reset limiter
	keepedTime := time.Duration(l.Bucket)*time.Second/time.Duration(l.Bandwidth) - elapsed
	if keepedTime > 0 {
		time.Sleep(keepedTime)
		l.KeepTime = keepedTime
		l.reset()
		return
	}

	// reset the limiter when stall threshold is smaller than elapsed time
	estimation := time.Duration(bufSize/l.Bandwidth) * time.Second
	stallThreshold := time.Second + estimation
	if elapsed > stallThreshold {
		l.reset()
	}
}
