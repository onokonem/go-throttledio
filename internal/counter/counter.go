package counter

import (
	"math"
	"sync"
	"time"
)

var maxFloat64Int64 = func() float64 { // nolint: gochecknoglobals
	for i := int64(math.MaxInt64); i > 0; i-- {
		if int64(float64(i)) == i {
			return float64(i)
		}
	}
	panic("unreachable reached")
}()

// Counter used to stack up the measures back to the defined period of time.
// Old measureas are discarded.
type Counter struct {
	mtime            time.Time
	tickDuration     time.Duration
	intervalDuration time.Duration
	tick             int
	counts           []int64
	lock             sync.Mutex
}

// NewCounter creates a counter.
// interval is a perion of time the measuring performed.
// ticks is a number of time gaps interval will be divided to.
// More ticks mean more accuracy.
func NewCounter(interval time.Duration, ticks uint) *Counter {
	curTime := time.Now()
	tickDuration := interval / time.Duration(ticks)
	return &Counter{
		mtime:            curTime.Truncate(tickDuration),
		intervalDuration: interval,
		tickDuration:     tickDuration,
		tick:             int(curTime.Sub(curTime.Truncate(interval)) / tickDuration),
		counts:           make([]int64, ticks),
	}
}

// FillUp is used to add a measure. Returns a summ of al the measures passed back to the configured interval.
func (c *Counter) FillUp(n int64) int64 {
	c.lock.Lock()
	defer c.lock.Unlock()

	total := c.cleanUpLocked()
	c.counts[c.tick] += n

	return total + n
}

// FillUpToCap adding a measure to make total/interval ratio no bigger than cps (counts per second).
// Returns an actual amount was added.
func (c *Counter) FillUpToCap(n int64, cps int64) int64 {
	maxByCPS := float64(cps) * c.intervalDuration.Seconds()
	if maxByCPS > maxFloat64Int64 {
		maxByCPS = maxFloat64Int64
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	left := int64(maxByCPS) - c.cleanUpLocked()

	switch {
	case left <= 0:
		return 0
	case left >= n:
		c.counts[c.tick] += n
		return n
	}

	c.counts[c.tick] += left
	return left
}

func (c *Counter) cleanUpLocked() int64 {
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

// Reset the measures to be used with new CPS
func (c *Counter) Reset(cps int64) {
	v := math.MaxInt64 / int64(len(c.counts))

	if maxByCPS := float64(cps) * c.intervalDuration.Seconds(); maxByCPS <= maxFloat64Int64 {
		v = int64(maxByCPS) / int64(len(c.counts))
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	for i := range c.counts {
		c.counts[i] = v
	}

	curTime := time.Now()
	c.mtime = curTime.Truncate(c.tickDuration)
	c.tick = int(curTime.Sub(curTime.Truncate(c.intervalDuration)) / c.tickDuration)
	c.counts[c.tick] = 0
}
