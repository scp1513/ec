package parallel

import (
	"log"
	"runtime"
	"sync"
)

var (
	goFuncRecv   = true
	goFuncNum    int
	goFuncMutex  sync.Mutex
	goFuncFinish chan struct{}
	goRecover    = defRecover
)

func SetGORecv(v bool) {
	goFuncRecv = v
}

func SetRecover(f func(interface{})) {
	goRecover = f
}

func defRecover(e interface{}) {
	var buf [4096]byte
	i := runtime.Stack(buf[:], false)
	log.Printf("%#v\n%s\n", e, buf[:i])
}

func GO(f func()) {
	if !goFuncRecv {
		return
	}
	goInc()
	go func() {
		invoke(f)
		goDec()
	}()
}

func invoke(fn func()) {
	if goRecover != nil {
		defer func() {
			if e := recover(); e != nil {
				goRecover(e)
			}
		}()
	}
	fn()
}

func goInc() {
	goFuncMutex.Lock()
	goFuncNum++
	goFuncMutex.Unlock()
}

func goDec() {
	goFuncMutex.Lock()
	goFuncNum--
	if goFuncNum == 0 {
		goNotifyFinish()
	}
	goFuncMutex.Unlock()
}

func goNotifyFinish() {
	if goFuncFinish != nil {
		goFuncFinish <- struct{}{}
	}
}

// WaitGO 等待所有go结束，多次调用只有一次能够接收channal
func WaitGO() <-chan struct{} {
	goFuncMutex.Lock()
	if goFuncFinish == nil {
		goFuncFinish = make(chan struct{}, 1)
	}
	if goFuncNum == 0 {
		goFuncFinish <- struct{}{}
	}
	goFuncMutex.Unlock()
	return goFuncFinish
}
