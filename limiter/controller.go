package limiter

import (
	"errors"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/onokonem/go-throttledio/internal/counter"
)

// Errors
var (
	ErrInvalidParams = errors.New("parameter invalid")
)

// Controller is a struct to create and control Limiters.
type Controller struct {
	interval    time.Duration
	ticks       uint
	counter     *counter.Counter
	commonCPS   int64
	perChildCPS int64
}

// NewController creates a new controller instance.
// interval is a period of time measuring is performed.
// ticks is a number of gaps interval is divided to.
// commonCPS is a spped liimit common for all the derrived Limiters.
// perChildCPS is a default value for each derrived Limiter.
// Note: each Limiter have 3 limits: it's own one, perChildCPS, and commonCPS. Minimal one is applied in any case.
// commonCPS is a limit for all the derrived Limiters together.
func NewController(interval time.Duration, ticks uint, commonCPS int64, perChildCPS int64) *Controller {
	if commonCPS <= 0 {
		commonCPS = math.MaxInt64
	}

	if perChildCPS <= 0 {
		perChildCPS = math.MaxInt64
	}

	if ticks == 0 {
		panic(fmt.Errorf("ticks: %d: %w", ticks, ErrInvalidParams))
	}

	c := &Controller{
		interval:    interval,
		ticks:       ticks,
		counter:     counter.NewCounter(interval, ticks),
		commonCPS:   commonCPS,
		perChildCPS: perChildCPS,
	}

	c.counter.Reset(commonCPS)

	return c
}

// BornLimiter returns a new limiter
func (c *Controller) BornLimiter() *Limiter {
	l := &Limiter{
		controller: c,
		counter:    counter.NewCounter(c.interval, c.ticks),
		cps:        atomic.LoadInt64(&c.perChildCPS),
	}

	l.counter.Reset(l.cps)

	return l
}

// SetCommonCPS sets the one for all together limit
func (c *Controller) SetCommonCPS(cps int64) {
	if cps <= 0 {
		cps = math.MaxInt64
	}

	atomic.StoreInt64(&c.commonCPS, cps)
	c.counter.Reset(cps)
}

// SetPerChildCPS changes a default limit for the new Limiters and also reset the limit on al the existing Limiters.
func (c *Controller) SetPerChildCPS(cps int64) {
	if cps <= 0 {
		cps = math.MaxInt64
	}

	atomic.StoreInt64(&c.perChildCPS, cps)
}
