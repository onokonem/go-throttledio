package limiter_test

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/onokonem/go-throttledio/limiter"
	"golang.org/x/xerrors"
)

// u u u
// u u b
// u b u
// b u u
// s m b
// b m s
// b s m
// m b s

const (
	interval     = time.Second * 3
	ticks        = 100
	concurency   = 10
	testDuration = interval * 7 / 2
	maxDeviation = 0.05
)

func init() { rand.Seed(time.Now().UnixNano()) }

func fillUp(l *limiter.Limiter, endTime time.Time, sleep time.Duration, total *int64, wg *sync.WaitGroup) {
	defer wg.Done()

	for ; time.Now().Before(endTime); time.Sleep(sleep) {
		n := rand.Int63n(math.MaxInt64 / 1000000000)
		a := l.FillUp(n)
		atomic.AddInt64(total, a)
	}
}

func calculateResult(startTime, endTime time.Time, total int64, reqCPS float64) (time.Duration, float64, float64) {
	var (
		spent     = endTime.Sub(startTime)
		actualCPS = float64(total) / spent.Seconds()
		deviation = (reqCPS - actualCPS) / reqCPS
	)
	return spent, actualCPS, deviation
}

func TestUUU(t *testing.T) {
	var (
		total     = int64(0)
		startTime = time.Now()
		endTime   = time.Now().Add(testDuration)
		limiters  = make([]*limiter.Limiter, concurency)
		reqCPS    = int64(0)
		c         = limiter.NewController(interval, ticks, 0, 0)
	)

	var wg sync.WaitGroup
	wg.Add(concurency)
	for ri := 0; ri < concurency; ri++ {
		limiters[ri] = c.BornLimiter()
		go fillUp(limiters[ri], endTime, testDuration/ticks/2, &total, &wg)
	}
	wg.Wait()

	spent, actualCPS, deviation := calculateResult(startTime, time.Now(), total, float64(reqCPS))
	fmt.Printf("counted: %d, spent: %s, cps: %dx%d, actual: %3.3f, deviation: %3.3f\n", total, spent, reqCPS, concurency, actualCPS, deviation)
}

func TestUUB(t *testing.T) {
	var (
		total     = int64(0)
		startTime = time.Now()
		endTime   = time.Now().Add(testDuration)
		limiters  = make([]*limiter.Limiter, concurency)
		reqCPS    = rand.Int63n(1000) + 1000
		c         = limiter.NewController(interval, ticks, 0, 0)
	)

	var wg sync.WaitGroup
	wg.Add(concurency)
	for ri := 0; ri < concurency; ri++ {
		limiters[ri] = c.BornLimiter()
		go fillUp(limiters[ri], endTime, testDuration/ticks/2, &total, &wg)
	}

	time.Sleep(testDuration / 2)

	halfTotal := atomic.LoadInt64(&total)
	spent, actualCPS, deviation := calculateResult(startTime, time.Now(), halfTotal, 0)
	fmt.Printf("counted: %d, spent: %s, cps: %dx%d, actual: %3.3f, deviation: %3.3f\n", halfTotal, spent, reqCPS, concurency, actualCPS, deviation)

	for _, l := range limiters {
		l.SetCPS(reqCPS)
	}
	atomic.StoreInt64(&total, 0)
	startTime = time.Now()

	wg.Wait()

	spent, actualCPS, deviation = calculateResult(startTime, time.Now(), total, float64(reqCPS*concurency))
	fmt.Printf("counted: %d, spent: %s, cps: %dx%d, actual: %3.3f, deviation: %3.3f\n", total, spent, reqCPS, concurency, actualCPS, deviation)
	if math.Abs(deviation) > maxDeviation {
		t.Errorf("deviation is too big: %3.3f > %3.3f", deviation, maxDeviation)
	}
}

func TestUBU1(t *testing.T) {
	var (
		total     = int64(0)
		startTime = time.Now()
		endTime   = time.Now().Add(testDuration)
		limiters  = make([]*limiter.Limiter, concurency)
		reqCPS    = rand.Int63n(1000) + 1000
		c         = limiter.NewController(interval, ticks, 0, 0)
	)

	var wg sync.WaitGroup
	wg.Add(concurency)
	for ri := 0; ri < concurency; ri++ {
		limiters[ri] = c.BornLimiter()
		go fillUp(limiters[ri], endTime, testDuration/ticks/2, &total, &wg)
	}

	time.Sleep(testDuration / 2)

	halfTotal := atomic.LoadInt64(&total)
	spent, actualCPS, deviation := calculateResult(startTime, time.Now(), halfTotal, float64(reqCPS*concurency))
	fmt.Printf("counted: %d, spent: %s, cps: %dx%d, actual: %3.3f, deviation: %3.3f\n", halfTotal, spent, reqCPS, concurency, actualCPS, deviation)

	c.SetPerChildCPS(reqCPS)
	atomic.StoreInt64(&total, 0)
	startTime = time.Now()

	wg.Wait()

	spent, actualCPS, deviation = calculateResult(startTime, time.Now(), total, float64(reqCPS*concurency))
	fmt.Printf("counted: %d, spent: %s, cps: %dx%d, actual: %3.3f, deviation: %3.3f\n", total, spent, reqCPS, concurency, actualCPS, deviation)
	if math.Abs(deviation) > maxDeviation {
		t.Errorf("deviation is too big: %3.3f > %3.3f", deviation, maxDeviation)
	}
}

func TestUBU2(t *testing.T) {
	var (
		total     = int64(0)
		startTime = time.Now()
		endTime   = time.Now().Add(testDuration)
		limiters  = make([]*limiter.Limiter, concurency)
		reqCPS    = rand.Int63n(1000) + 1000
		c         = limiter.NewController(interval, ticks, 1, reqCPS)
	)

	c.SetCommonCPS(0)

	var wg sync.WaitGroup
	wg.Add(concurency)
	for ri := 0; ri < concurency; ri++ {
		limiters[ri] = c.BornLimiter()
		go fillUp(limiters[ri], endTime, testDuration/ticks/2, &total, &wg)
	}

	time.Sleep(testDuration / 2)

	halfTotal := atomic.LoadInt64(&total)
	spent, actualCPS, deviation := calculateResult(startTime, time.Now(), halfTotal, float64(reqCPS*concurency))
	fmt.Printf("counted: %d, spent: %s, cps: %dx%d, actual: %3.3f, deviation: %3.3f\n", halfTotal, spent, reqCPS, concurency, actualCPS, deviation)
	if math.Abs(deviation) > maxDeviation {
		t.Errorf("deviation is too big: %3.3f > %3.3f", deviation, maxDeviation)
	}

	c.SetPerChildCPS(reqCPS)
	atomic.StoreInt64(&total, 0)
	startTime = time.Now()

	wg.Wait()

	spent, actualCPS, deviation = calculateResult(startTime, time.Now(), total, float64(reqCPS*concurency))
	fmt.Printf("counted: %d, spent: %s, cps: %dx%d, actual: %3.3f, deviation: %3.3f\n", total, spent, reqCPS, concurency, actualCPS, deviation)
	if math.Abs(deviation) > maxDeviation {
		t.Errorf("deviation is too big: %3.3f > %3.3f", deviation, maxDeviation)
	}
}

func TestBUU1(t *testing.T) {
	var (
		total     = int64(0)
		startTime = time.Now()
		endTime   = time.Now().Add(testDuration)
		limiters  = make([]*limiter.Limiter, concurency)
		reqCPS    = rand.Int63n(1000) + 1000
		c         = limiter.NewController(interval, ticks, 0, 0)
	)

	var wg sync.WaitGroup
	wg.Add(concurency)
	for ri := 0; ri < concurency; ri++ {
		limiters[ri] = c.BornLimiter()
		limiters[ri].SetCPS(0)
		go fillUp(limiters[ri], endTime, testDuration/ticks/2, &total, &wg)
	}

	time.Sleep(testDuration / 2)

	halfTotal := atomic.LoadInt64(&total)
	spent, actualCPS, deviation := calculateResult(startTime, time.Now(), halfTotal, float64(0))
	fmt.Printf("counted: %d, spent: %s, cps: %dx%d, actual: %3.3f, deviation: %3.3f\n", halfTotal, spent, reqCPS, concurency, actualCPS, deviation)

	c.SetCommonCPS(reqCPS)
	atomic.StoreInt64(&total, 0)
	startTime = time.Now()

	wg.Wait()

	spent, actualCPS, deviation = calculateResult(startTime, time.Now(), total, float64(reqCPS))
	fmt.Printf("counted: %d, spent: %s, cps: %dx%d, actual: %3.3f, deviation: %3.3f\n", total, spent, reqCPS, concurency, actualCPS, deviation)
	if math.Abs(deviation) > maxDeviation {
		t.Errorf("deviation is too big: %3.3f > %3.3f", deviation, maxDeviation)
	}
}

func TestBUU2(t *testing.T) {
	var (
		total     = int64(0)
		startTime = time.Now()
		endTime   = time.Now().Add(testDuration)
		limiters  = make([]*limiter.Limiter, concurency)
		reqCPS    = rand.Int63n(1000) + 1000
		c         = limiter.NewController(interval, ticks, reqCPS, 1)
	)

	c.SetPerChildCPS(0)

	var wg sync.WaitGroup
	wg.Add(concurency)
	for ri := 0; ri < concurency; ri++ {
		limiters[ri] = c.BornLimiter()
		go fillUp(limiters[ri], endTime, testDuration/ticks/2, &total, &wg)
	}

	time.Sleep(testDuration / 2)

	halfTotal := atomic.LoadInt64(&total)
	spent, actualCPS, deviation := calculateResult(startTime, time.Now(), halfTotal, float64(reqCPS))
	fmt.Printf("counted: %d, spent: %s, cps: %dx%d, actual: %3.3f, deviation: %3.3f\n", halfTotal, spent, reqCPS, concurency, actualCPS, deviation)
	if math.Abs(deviation) > maxDeviation {
		t.Errorf("deviation is too big: %3.3f > %3.3f", deviation, maxDeviation)
	}

	c.SetCommonCPS(reqCPS)
	atomic.StoreInt64(&total, 0)
	startTime = time.Now()

	wg.Wait()

	spent, actualCPS, deviation = calculateResult(startTime, time.Now(), total, float64(reqCPS))
	fmt.Printf("counted: %d, spent: %s, cps: %dx%d, actual: %3.3f, deviation: %3.3f\n", total, spent, reqCPS, concurency, actualCPS, deviation)
	if math.Abs(deviation) > maxDeviation {
		t.Errorf("deviation is too big: %3.3f > %3.3f", deviation, maxDeviation)
	}
}

func TestSMB(t *testing.T) {
	var (
		total     = int64(0)
		startTime = time.Now()
		endTime   = time.Now().Add(testDuration)
		limiters  = make([]*limiter.Limiter, concurency)
		reqCPSs   = limit{name: "common", concurency: 1, cps: rand.Int63n(1000) + 1000}
		reqCPSm   = limit{name: "perChild", concurency: concurency, cps: reqCPSs.cps * (rand.Int63n(20) + 2)}
		reqCPSb   = limit{name: "limiter", concurency: concurency, cps: reqCPSm.cps * (rand.Int63n(20) + 2)}
	)
	c := limiter.NewController(interval, ticks, reqCPSs.cps, reqCPSm.cps)

	var wg sync.WaitGroup
	wg.Add(concurency)
	for ri := 0; ri < concurency; ri++ {
		limiters[ri] = c.BornLimiter()
		limiters[ri].SetCPS(reqCPSb.cps)
		go fillUp(limiters[ri], endTime, testDuration/ticks/2, &total, &wg)
	}

	wg.Wait()

	reqCPS, lim := minLimit(reqCPSs, reqCPSm, reqCPSb)

	spent, actualCPS, deviation := calculateResult(startTime, time.Now(), total, float64(reqCPS))
	fmt.Printf("counted: %d, spent: %s, cps %s: %dx%d, actual: %3.3f, deviation: %3.3f\n", total, spent, lim.name, lim.cps, lim.concurency, actualCPS, deviation)
	if math.Abs(deviation) > maxDeviation {
		t.Errorf("deviation is too big: %3.3f > %3.3f", deviation, maxDeviation)
	}
}

func TestBMS(t *testing.T) {
	var (
		total     = int64(0)
		startTime = time.Now()
		endTime   = time.Now().Add(testDuration)
		limiters  = make([]*limiter.Limiter, concurency)
		reqCPSs   = limit{name: "limiter", concurency: concurency, cps: rand.Int63n(1000) + 1000}
		reqCPSm   = limit{name: "perChild", concurency: concurency, cps: reqCPSs.cps * (rand.Int63n(20) + 2)}
		reqCPSb   = limit{name: "common", concurency: 1, cps: reqCPSm.cps * (rand.Int63n(20) + 2)}
	)
	c := limiter.NewController(interval, ticks, reqCPSb.cps, reqCPSm.cps)

	var wg sync.WaitGroup
	wg.Add(concurency)
	for ri := 0; ri < concurency; ri++ {
		limiters[ri] = c.BornLimiter()
		limiters[ri].SetCPS(reqCPSs.cps)
		go fillUp(limiters[ri], endTime, testDuration/ticks/2, &total, &wg)
	}

	wg.Wait()

	reqCPS, lim := minLimit(reqCPSs, reqCPSm, reqCPSb)

	spent, actualCPS, deviation := calculateResult(startTime, time.Now(), total, float64(reqCPS))
	fmt.Printf("counted: %d, spent: %s, cps %s: %dx%d, actual: %3.3f, deviation: %3.3f\n", total, spent, lim.name, lim.cps, lim.concurency, actualCPS, deviation)
	if math.Abs(deviation) > maxDeviation {
		t.Errorf("deviation is too big: %3.3f > %3.3f", deviation, maxDeviation)
	}
}

func TestBSM(t *testing.T) {
	var (
		total     = int64(0)
		startTime = time.Now()
		endTime   = time.Now().Add(testDuration)
		limiters  = make([]*limiter.Limiter, concurency)
		reqCPSs   = limit{name: "perChild", concurency: concurency, cps: rand.Int63n(1000) + 1000}
		reqCPSm   = limit{name: "limiter", concurency: concurency, cps: reqCPSs.cps * (rand.Int63n(20) + 2)}
		reqCPSb   = limit{name: "common", concurency: 1, cps: reqCPSm.cps * (rand.Int63n(20) + 2)}
	)
	c := limiter.NewController(interval, ticks, reqCPSb.cps, reqCPSs.cps)

	var wg sync.WaitGroup
	wg.Add(concurency)
	for ri := 0; ri < concurency; ri++ {
		limiters[ri] = c.BornLimiter()
		limiters[ri].SetCPS(reqCPSm.cps)
		go fillUp(limiters[ri], endTime, testDuration/ticks/2, &total, &wg)
	}

	wg.Wait()

	reqCPS, lim := minLimit(reqCPSs, reqCPSm, reqCPSb)

	spent, actualCPS, deviation := calculateResult(startTime, time.Now(), total, float64(reqCPS))
	fmt.Printf("counted: %d, spent: %s, cps %s: %dx%d, actual: %3.3f, deviation: %3.3f\n", total, spent, lim.name, lim.cps, lim.concurency, actualCPS, deviation)
	if math.Abs(deviation) > maxDeviation {
		t.Errorf("deviation is too big: %3.3f > %3.3f", deviation, maxDeviation)
	}
}

func TestMBS(t *testing.T) {
	var (
		total     = int64(0)
		startTime = time.Now()
		endTime   = time.Now().Add(testDuration)
		limiters  = make([]*limiter.Limiter, concurency)
		reqCPSs   = limit{name: "limiter", concurency: concurency, cps: rand.Int63n(1000) + 1000}
		reqCPSm   = limit{name: "common", concurency: 1, cps: reqCPSs.cps * (rand.Int63n(20) + 2)}
		reqCPSb   = limit{name: "perChild", concurency: concurency, cps: reqCPSm.cps * (rand.Int63n(20) + 2)}
	)
	c := limiter.NewController(interval, ticks, reqCPSm.cps, reqCPSb.cps)

	var wg sync.WaitGroup
	wg.Add(concurency)
	for ri := 0; ri < concurency; ri++ {
		limiters[ri] = c.BornLimiter()
		limiters[ri].SetCPS(reqCPSs.cps)
		go fillUp(limiters[ri], endTime, testDuration/ticks/2, &total, &wg)
	}

	wg.Wait()

	reqCPS, lim := minLimit(reqCPSs, reqCPSm, reqCPSb)

	spent, actualCPS, deviation := calculateResult(startTime, time.Now(), total, float64(reqCPS))
	fmt.Printf("counted: %d, spent: %s, cps %s: %dx%d, actual: %3.3f, deviation: %3.3f\n", total, spent, lim.name, lim.cps, lim.concurency, actualCPS, deviation)
	if math.Abs(deviation) > maxDeviation {
		t.Errorf("deviation is too big: %3.3f > %3.3f", deviation, maxDeviation)
	}
}

func TestInvalidTick(t *testing.T) {
	e := testPanic(
		func() {
			limiter.NewController(interval, 0, 0, 0)
		},
	)

	fmt.Printf("got: %#+v\n", e)

	if err, ok := e.(error); !ok || !xerrors.Is(err, limiter.ErrInvalidParams) {
		t.Errorf("expected %#+v, got %#+v", limiter.ErrInvalidParams, e)
	}
}

func TestReturnCount(t *testing.T) {
	c := limiter.NewController(interval, ticks, 1, 0)

	l := c.BornLimiter()

	a := l.FillUp(10)
	expected := int64(interval.Seconds())
	if expected != a {
		t.Errorf("expected %#+v, got %#+v", expected, a)
	}

	a = l.FillUp(-100)
	expected = -100
	if expected != a {
		t.Errorf("expected %#+v, got %#+v", expected, a)
	}

	a = l.FillUp(1000)
	expected = 100
	if expected != a {
		t.Errorf("expected %#+v, got %#+v", expected, a)
	}
}

type limit struct {
	name       string
	cps        int64
	concurency int
}

func minLimit(vv ...limit) (int64, limit) {
	min := int64(math.MaxInt64)
	var lim limit

	for _, v := range vv {
		total := v.cps * int64(v.concurency)
		if min > total {
			min = total
			lim = v
		}
	}

	return min, lim
}

func testPanic(f func()) (e interface{}) {
	defer func() {
		e = recover()
	}()
	f()
	return nil
}
