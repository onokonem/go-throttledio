package counter_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/onokonem/go-throttledio/internal/counter"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	testDuration = time.Second * 10
	interval     = time.Second * 3
	ticks        = 10
	concurency   = 10
)

type testTick struct {
	t time.Time
	n int64
}

type testTicks struct {
	ticks []testTick
	lock  sync.Mutex
}

func (s *testTicks) append(v testTick) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ticks = append(s.ticks, v)
}

func TestFillUpRandom(t *testing.T) {
	toCheck := &testTicks{ticks: make([]testTick, 0, 500)}

	c := counter.NewCounter(interval, ticks)

	wg := sync.WaitGroup{}
	wg.Add(concurency)
	for i := 0; i < concurency; i++ {
		go func() {
			defer wg.Done()
			for end := time.Now().Add(testDuration); end.After(time.Now()); time.Sleep(randomDelay()) {
				n := int64(rand.Intn(1500) + 1)
				toCheck.append(testTick{time.Now().Truncate(interval / ticks), n})
				c.FillUp(n)
			}
		}()
	}

	wg.Wait()

	curTime := time.Now()
	counted := c.FillUp(0)
	margin := curTime.Add(-interval).Truncate(interval / ticks)
	actual := int64(0)
	inInterval := 0
	for _, v := range toCheck.ticks {
		if v.t.After(margin) {
			inInterval++
			actual += v.n
		}
	}

	if counted != actual {
		t.Errorf("expected %d, got %d", actual, counted)
		return
	}

	fmt.Printf("success on %d counts, %d in interval\n", len(toCheck.ticks), inInterval)
}

func TestFillUpShort(t *testing.T) {
	toCheck := &testTicks{ticks: make([]testTick, 0, 500)}

	c := counter.NewCounter(interval, ticks)

	wg := sync.WaitGroup{}
	wg.Add(concurency)
	for i := 0; i < concurency; i++ {
		go func() {
			defer wg.Done()
			for end := time.Now().Add(interval * 3 / 2); end.After(time.Now()); time.Sleep(interval / ticks * 3 / 4) {
				n := int64(rand.Intn(1500) + 1)
				toCheck.append(testTick{time.Now().Truncate(interval / ticks), n})
				c.FillUp(n)
			}
		}()
	}

	wg.Wait()

	curTime := time.Now()
	counted := c.FillUp(0)
	margin := curTime.Add(-interval).Truncate(interval / ticks)
	actual := int64(0)
	inInterval := 0
	for _, v := range toCheck.ticks {
		if v.t.After(margin) {
			inInterval++
			actual += v.n
		}
	}

	if counted != actual {
		t.Errorf("expected %d, got %d", actual, counted)
		return
	}

	fmt.Printf("success on %d counts, %d in interval\n", len(toCheck.ticks), inInterval)
}

func TestFillUpLong(t *testing.T) {
	c := counter.NewCounter(interval, ticks)

	for end := time.Now().Add(interval * 3 / 2); end.After(time.Now()); time.Sleep(interval / ticks * 3 / 4) {
		c.FillUp(int64(int64(rand.Intn(1500) + 1)))
	}

	time.Sleep(interval)
	actual := int64(rand.Intn(1500) + 1)
	counted := c.FillUp(int64(actual))

	if counted != actual {
		t.Errorf("expected %d, got %d", actual, counted)
		return
	}
}

// 10%  0-5ms
// 85%  30-100ms
// 5%  1000-3500ms
const (
	smallPercent  = 10
	smallMin      = 0
	smallMax      = 3
	mediumPercent = 85
	mediumMin     = 10
	mediumMax     = 30
	largePercent  = 100 - smallPercent - smallPercent // useless
	largeMin      = 1000
	largeMax      = 3500
)

func randomDelay() time.Duration {
	x := rand.Intn(100)
	switch {
	case x < smallPercent:
		return time.Duration(rand.Intn(smallMax-smallMin)+smallMin) * time.Millisecond
	case x < (smallPercent + mediumPercent):
		return time.Duration(rand.Intn(mediumMax-mediumMin)+mediumMin) * time.Millisecond
	default:
		return time.Duration(rand.Intn(largeMax-largeMin)+largeMin) * time.Millisecond
	}
}

func BenchmarkFillUp(b *testing.B) {
	c := counter.NewCounter(interval, ticks)
	for i := 0; i < b.N; i++ {
		c.FillUp(1)
	}
}
