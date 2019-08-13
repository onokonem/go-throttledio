package counter

import (
	"sync"
	"time"
)

// Counter used to stack up the measures back to the defined period of time.
// Old measureas are discarded.
type Counter struct {
	mtime        time.Time
	tickDuration time.Duration
	tick         int
	counts       []int64
	lock         sync.Mutex
}

// NewCounter creates a counter.
// interval is a perion of time the measuring performed.
// ticks is a number of time gaps interval will be divided to.
// More ticks mean more acuracy.
func NewCounter(interval time.Duration, ticks uint) *Counter {
	return &Counter{
		mtime:        time.Now().Truncate(interval),
		tickDuration: interval / time.Duration(ticks),
		counts:       make([]int64, ticks),
	}
}

// FillUp is used to add a measure. Returns a summ of al the measures passed back to the configured interval.
func (c *Counter) FillUp(n int64) int64 {
	c.lock.Lock()
	defer c.lock.Unlock()

	var (
		curTime = time.Now()
		gap     = int(curTime.Sub(c.mtime) / c.tickDuration)
	)

	if gap >= len(c.counts) {
		gap = len(c.counts)
	}

	for i := c.tick + gap; i > c.tick; i-- {
		c.counts[i%len(c.counts)] = 0
	}

	c.tick = (c.tick + gap) % len(c.counts)
	c.counts[c.tick] += n
	c.mtime = curTime.Truncate(c.tickDuration)

	return c.totalLocked()
}

func (c *Counter) totalLocked() int64 {
	total := int64(0)
	for _, v := range c.counts {
		total += v
	}

	return total
}
