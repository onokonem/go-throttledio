package counter_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/onokonem/go-throttledio/internal/counter"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	testDuration = time.Second * 25
	interval     = time.Second * 3
	ticks        = 10
)

func TestFillUp(t *testing.T) {
	toCheck := make(map[time.Time]int64, 200)

	c := counter.NewCounter(interval, ticks)
	counted := int64(0)

	curTime := time.Now()
	for end := time.Now().Add(testDuration); end.After(time.Now()); time.Sleep(randomDelay()) {
		n := int64(rand.Intn(1500) + 1)
		curTime = time.Now()
		counted = c.FillUp(int64(n))
		toCheck[time.Now()] += n
	}

	margin := curTime.Add(-interval)
	actual := int64(0)
	inInterval := 0
	for k, v := range toCheck {
		if k.After(margin) {
			inInterval++
			actual += v
		}
	}

	if counted != actual {
		t.Errorf("expected %d, got %d", actual, counted)
	}

	fmt.Printf("success on %d counts, %d in interval\n", len(toCheck), inInterval)
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

func BenchmarkFoo(b *testing.B) {
	c := counter.NewCounter(interval, ticks)
	for i := 0; i < b.N; i++ {
		c.FillUp(1)
	}
}
