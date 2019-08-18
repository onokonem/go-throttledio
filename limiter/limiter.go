package limiter

import (
	"math"
	"sync/atomic"

	"github.com/onokonem/go-throttledio/internal/counter"
)

// Limiter is an actual limiting unit
type Limiter struct {
	controller  *Controller
	counter     *counter.Counter
	cps         int64
	perChildCPS int64
}

// SetCPS sets the limit.
func (l *Limiter) SetCPS(cps int64) {
	if cps <= 0 {
		cps = math.MaxInt64
	}

	atomic.StoreInt64(&l.cps, cps)
	l.counter.Reset(minInt64(cps, atomic.LoadInt64(&l.controller.perChildCPS)))
}

// FillUp is used to report counter to Limiter.
func (l *Limiter) FillUp(n int64) int64 {
	switch {
	case n == 0:
		return n
	case n < 0:
		l.counter.FillUp(n)
		l.controller.counter.FillUp(n)
		return n
	}

	perChildCPS := atomic.LoadInt64(&l.controller.perChildCPS)
	cps := minInt64(atomic.LoadInt64(&l.cps), perChildCPS)

	if l.perChildCPS != perChildCPS {
		l.perChildCPS = perChildCPS
		l.counter.Reset(cps)
	}

	allowed := l.counter.FillUpToCap(n, cps)
	allowedCommon := l.controller.counter.FillUpToCap(allowed, atomic.LoadInt64(&l.controller.commonCPS))

	if allowedCommon < allowed {
		l.counter.FillUp(allowedCommon - allowed)
		return allowedCommon
	}

	return allowed
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
