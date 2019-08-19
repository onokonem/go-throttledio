package netlisten_test

import (
	"fmt"
	"math"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/onokonem/go-throttledio/netlisten"
)

// setting bandwidth limit per server
// setting bandwidth limit per connection
// changing limits in runtime (applies to all existing connections)
// for a 30s transfer sample consumed bandwidth should be accurate +/- 5%

const (
	MaxDeviation = 0.05
	Interval     = 10 * time.Second
	Ticks        = 100
	TestDuration = 30 * time.Second
)

// setting bandwidth limit per server
func TestPerServer(t *testing.T) {
	l := listen()

	clientG := new(int64)
	client1, client2 := startClients(l.Addr(), clientG)

	serverG := new(int64)
	startAccept(l, serverG)

	startTime := time.Now()

	time.Sleep(Interval / 3)

	curTime := time.Now()
	speed1 := float64(atomic.LoadInt64(client1.total)) / curTime.Sub(startTime).Seconds()
	speed2 := float64(atomic.LoadInt64(client2.total)) / curTime.Sub(startTime).Seconds()
	speedC := float64(atomic.LoadInt64(clientG)) / curTime.Sub(startTime).Seconds()

	fmt.Printf("Unlimited: speed 1: %f, speed 2: %f, common: %f\n", speed1, speed2, speedC)

	// setting bandwidth limit per server
	reqSpeedC := int64(speedC) / 10
	l.(*netlisten.Listener).ReadLimiter().SetCommonCPS(reqSpeedC)
	atomic.StoreInt64(client1.total, 0)
	atomic.StoreInt64(client2.total, 0)
	atomic.StoreInt64(clientG, 0)

	startTime = time.Now()
	time.Sleep(TestDuration)

	curTime = time.Now()
	speed1 = float64(atomic.LoadInt64(client1.total)) / curTime.Sub(startTime).Seconds()
	speed2 = float64(atomic.LoadInt64(client2.total)) / curTime.Sub(startTime).Seconds()
	speedC = float64(atomic.LoadInt64(clientG)) / curTime.Sub(startTime).Seconds()

	deviationC := (speedC - float64(reqSpeedC)) / float64(reqSpeedC)

	fmt.Printf("Limits: common %d: speed 1: %f, speed 2: %f, common: %f (d: %3.3f) \n", reqSpeedC, speed1, speed2, speedC, deviationC)

	if math.Abs(deviationC) > MaxDeviation {
		t.Errorf("common deviation too big: %3.3f", deviationC)
	}
}

// setting bandwidth limit per connection
func TestPerConn(t *testing.T) {
	l := listen()

	clientG := new(int64)
	client1, client2 := startClients(l.Addr(), clientG)

	serverG := new(int64)
	server1, server2 := startAccept(l, serverG)

	startTime := time.Now()

	time.Sleep(Interval / 3)

	curTime := time.Now()
	speed1 := float64(atomic.LoadInt64(client1.total)) / curTime.Sub(startTime).Seconds()
	speed2 := float64(atomic.LoadInt64(client2.total)) / curTime.Sub(startTime).Seconds()
	speedC := float64(atomic.LoadInt64(clientG)) / curTime.Sub(startTime).Seconds()

	fmt.Printf("Unlimited: speed 1: %f, speed 2: %f, common: %f\n", speed1, speed2, speedC)

	// setting bandwidth limit per connection
	reqSpeedC := int64(speedC) / 10
	reqSpeed1 := reqSpeedC / 2
	server1.conn.SetReadCPS(reqSpeed1)
	reqSpeed2 := reqSpeedC / 3
	server2.conn.SetReadCPS(reqSpeed2)
	atomic.StoreInt64(client1.total, 0)
	atomic.StoreInt64(client2.total, 0)
	atomic.StoreInt64(clientG, 0)

	startTime = time.Now()
	time.Sleep(TestDuration)

	curTime = time.Now()
	speed1 = float64(atomic.LoadInt64(client1.total)) / curTime.Sub(startTime).Seconds()
	speed2 = float64(atomic.LoadInt64(client2.total)) / curTime.Sub(startTime).Seconds()
	speedC = float64(atomic.LoadInt64(clientG)) / curTime.Sub(startTime).Seconds()

	deviation1 := (speed1 - float64(reqSpeed1)) / float64(reqSpeed1)
	deviation2 := (speed2 - float64(reqSpeed2)) / float64(reqSpeed2)

	fmt.Printf("Limits: 1 %d, 2 %d: speed 1: %f (d: %3.3f), speed 2: %f (d: %3.3f), common: %f\n", reqSpeed1, reqSpeed2, speed1, deviation1, speed2, deviation2, speedC)

	if math.Abs(deviation1) > MaxDeviation {
		t.Errorf("1 deviation too big: %3.3f", deviation1)
	}
	if math.Abs(deviation2) > MaxDeviation {
		t.Errorf("2 deviation too big: %3.3f", deviation2)
	}
}

// changing limits in runtime (applies to all existing connections)
func TestAllConn(t *testing.T) {
	l := listen()

	clientG := new(int64)
	client1, client2 := startClients(l.Addr(), clientG)

	serverG := new(int64)
	startAccept(l, serverG)

	startTime := time.Now()

	time.Sleep(Interval / 3)

	curTime := time.Now()
	speed1 := float64(atomic.LoadInt64(client1.total)) / curTime.Sub(startTime).Seconds()
	speed2 := float64(atomic.LoadInt64(client2.total)) / curTime.Sub(startTime).Seconds()
	speedC := float64(atomic.LoadInt64(clientG)) / curTime.Sub(startTime).Seconds()

	fmt.Printf("Unlimited: speed 1: %f, speed 2: %f, common: %f\n", speed1, speed2, speedC)

	// changing limits in runtime (applies to all existing connections)
	reqSpeedC := int64(speedC) / 10
	reqSpeed1 := reqSpeedC / 3
	reqSpeed2 := reqSpeedC / 3
	l.(*netlisten.Listener).ReadLimiter().SetPerChildCPS(reqSpeed1)
	atomic.StoreInt64(client1.total, 0)
	atomic.StoreInt64(client2.total, 0)
	atomic.StoreInt64(clientG, 0)

	startTime = time.Now()
	time.Sleep(TestDuration)

	curTime = time.Now()
	speed1 = float64(atomic.LoadInt64(client1.total)) / curTime.Sub(startTime).Seconds()
	speed2 = float64(atomic.LoadInt64(client2.total)) / curTime.Sub(startTime).Seconds()
	speedC = float64(atomic.LoadInt64(clientG)) / curTime.Sub(startTime).Seconds()

	deviation1 := (speed1 - float64(reqSpeed1)) / float64(reqSpeed1)
	deviation2 := (speed2 - float64(reqSpeed2)) / float64(reqSpeed2)

	fmt.Printf("Limits: 1 %d, 2 %d: speed 1: %f (d: %3.3f), speed 2: %f (d: %3.3f), common: %f\n", reqSpeed1, reqSpeed2, speed1, deviation1, speed2, deviation2, speedC)

	if math.Abs(deviation1) > MaxDeviation {
		t.Errorf("1 deviation too big: %3.3f", deviation1)
	}
	if math.Abs(deviation2) > MaxDeviation {
		t.Errorf("2 deviation too big: %3.3f", deviation2)
	}
}

func TestPerServerSetOnStart(t *testing.T) {
	l, err := net.Listen("tcp4", "127.0.0.1:")
	if err != nil {
		panic(err)
	}

	reqSpeedC := 10000000
	l = netlisten.LimitListener(l, reqSpeedC, reqSpeedC)

	clientG := new(int64)
	client1, client2 := startClients(l.Addr(), clientG)

	serverG := new(int64)
	startAccept(l, serverG)

	startTime := time.Now()

	time.Sleep(TestDuration)

	curTime := time.Now()
	speed1 := float64(atomic.LoadInt64(client1.total)) / curTime.Sub(startTime).Seconds()
	speed2 := float64(atomic.LoadInt64(client2.total)) / curTime.Sub(startTime).Seconds()
	speedC := float64(atomic.LoadInt64(clientG)) / curTime.Sub(startTime).Seconds()

	deviationC := (speedC - float64(reqSpeedC)) / float64(reqSpeedC)

	fmt.Printf("Limits: common %d: speed 1: %f, speed 2: %f, common: %f (d: %3.3f) \n", reqSpeedC, speed1, speed2, speedC, deviationC)

	if math.Abs(deviationC) > MaxDeviation {
		t.Errorf("common deviation too big: %3.3f", deviationC)
	}
}

type connReader struct {
	conn   *netlisten.Conn
	total  *int64
	global *int64
}

func (c *connReader) read() {
	b := make([]byte, 16*1024)
	for {
		n, err := c.conn.Read(b)
		if err != nil {
			panic(err)
		}

		atomic.AddInt64(c.total, int64(n))
		atomic.AddInt64(c.global, int64(n))
	}
}

func accept(l net.Listener, global *int64) *connReader {
	conn, err := l.Accept()
	if err != nil {
		panic(err)
	}

	return &connReader{
		conn:   conn.(*netlisten.Conn),
		total:  new(int64),
		global: global,
	}
}

type connWriter struct {
	total  *int64
	global *int64
}

func (c *connWriter) connectAndWrite(a net.Addr) {
	conn, err := net.Dial(a.Network(), a.String())
	if err != nil {
		panic(err)
	}

	b := make([]byte, 16*1024)
	for {
		n, err := conn.Write(b)
		if err != nil {
			panic(err)
		}

		atomic.AddInt64(c.total, int64(n))
		atomic.AddInt64(c.global, int64(n))
	}
}

func listen() net.Listener {
	l, err := net.Listen("tcp4", "127.0.0.1:")
	if err != nil {
		panic(err)
	}

	return netlisten.LimitListener(l, 0, 0)
}

func startClients(addr net.Addr, global *int64) (*connWriter, *connWriter) {
	return startClient(addr, global), startClient(addr, global)
}

func startClient(addr net.Addr, global *int64) *connWriter {
	client := &connWriter{total: new(int64), global: global}

	go client.connectAndWrite(addr)

	return client
}

func startAccept(l net.Listener, global *int64) (*connReader, *connReader) {
	server1 := accept(l, global)
	server2 := accept(l, global)

	go func() {
		conn, err := l.Accept()
		panic(fmt.Errorf("unexpected %v %v", conn, err))
	}()

	go server1.read()
	go server2.read()

	return server1, server2
}
