package serial_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"ec/serial"
)

func TestSerial(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	s := serial.NewSerial(4096)
	go s.Run()
	s.Expire(time.Second*0, func() { fmt.Println("expire") })
	s.ExpireAt(time.Now(), func() { fmt.Println("expire at") })
	s.Expire(time.Second*3, func() { fmt.Println("expire") })
	s.ExpireAt(time.Now().Add(time.Second*3), func() { fmt.Println("expire at") })
	s.AddTimer(time.Second, true, func() { fmt.Println("timer") })
	time.Sleep(time.Millisecond)
	for j := 0; j < 10; j++ {
		s.Post(func() {
			fmt.Println(j)
		})
	}
	fmt.Println("num timer", s.NumTimer())
	fmt.Println("num waiting", s.NumWaiting())
	time.Sleep(time.Second * 5)
	s.Stop()
	time.Sleep(time.Second * 2)
}

func TestStop(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	s := serial.NewSerial(4096)
	for i := 0; i < 10; i++ {
		n := new(int)
		*n = i + 1
		s.Post(func() { time.Sleep(time.Second); fmt.Println(*n) })
	}
	c := make(chan struct{}, 1)
	go func() {
		s.Run()
		c <- struct{}{}
	}()
	time.Sleep(time.Second)
	s.Stop()
	fmt.Println("stop")
	<-c
	fmt.Println("exit")
}
