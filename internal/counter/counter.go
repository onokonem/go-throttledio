package counter

import (
	"sync"
	"time"
)

type Counter struct {
	mtime        time.Time
	tickDuration time.Duration
	tick         int
	counts       []int64
	lock         sync.Mutex
}

func NewCounter(interval time.Duration, ticks uint) *Counter {

	return &Counter{
		mtime:        time.Now().Truncate(interval),
		tickDuration: interval / time.Duration(ticks),
		counts:       make([]int64, ticks),
	}
}

func (c *Counter) FillUp(n int64) int64 {
	c.lock.Lock()
	c.lock.Unlock()

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
