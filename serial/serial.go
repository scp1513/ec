// Package serial 串行化队列
package serial

import (
	"sync"
	"time"
)

// S 串行化队列
type S struct {
	handlerChan    chan func()
	timerIDMutex   sync.RWMutex
	timerIDCounter uint
	timerIDs       map[uint]chan struct{}
	recoverFunc    func(interface{})
}

// New 新建一个S
func New(size int) *S {
	return &S{handlerChan: make(chan func(), size), timerIDs: make(map[uint]chan struct{})}
}

func (s *S) SetRecoverFunc(f func(interface{})) {
	s.recoverFunc = f
}

// Start 开始队列
func (s *S) Start() {
	go s.Run()
}

// Run 开始队列
func (s *S) Run() {
	for fn := range s.handlerChan {
		s.invoke(fn)
	}
}

func (s *S) invoke(fn func()) {
	if s.recoverFunc != nil {
		defer func() {
			if e := recover(); e != nil {
				s.recoverFunc(e)
			}
		}()
	}
	fn()
}

// Stop 结束队列
func (s *S) Stop() {
	s.timerIDMutex.RLock()
	for k := range s.timerIDs {
		s.timerIDs[k] <- struct{}{}
	}
	s.timerIDMutex.RUnlock()
	close(s.handlerChan)
}

// Post 末端加入队列
func (s *S) Post(fn func()) {
	s.handlerChan <- fn
}

// ExpireAt 指定时间定时器
func (s *S) ExpireAt(t time.Time, fn func()) uint {
	if fn == nil {
		return 0
	}
	c := make(chan struct{}, 1)
	s.timerIDMutex.Lock()
	defer s.timerIDMutex.Unlock()

	s.timerIDCounter++
	s.timerIDs[s.timerIDCounter] = c

	go s.onceTimerHandler(fn, t, s.timerIDCounter)

	return s.timerIDCounter
}

// Expire 超时定时器
func (s *S) Expire(duration time.Duration, fn func()) uint {
	if fn == nil {
		return 0
	}
	t := time.Now().Add(duration)

	c := make(chan struct{}, 1)
	s.timerIDMutex.Lock()
	defer s.timerIDMutex.Unlock()

	s.timerIDCounter++
	s.timerIDs[s.timerIDCounter] = c

	go s.onceTimerHandler(fn, t, s.timerIDCounter)

	return s.timerIDCounter
}

// AddTimer 循环定时器
func (s *S) AddTimer(duration time.Duration, immediately bool, fn func()) uint {
	if fn == nil {
		return 0
	}
	now := time.Now()
	c := make(chan struct{}, 1)
	s.timerIDMutex.Lock()
	defer s.timerIDMutex.Unlock()

	s.timerIDCounter++
	s.timerIDs[s.timerIDCounter] = c

	go s.loopTimerHandler(fn, now, duration, s.timerIDCounter, immediately)

	return s.timerIDCounter
}

// CancelTimer 取消定时器
func (s *S) CancelTimer(timerID uint) {
	s.timerIDMutex.RLock()
	c, ok := s.timerIDs[timerID]
	s.timerIDMutex.RUnlock()

	if ok {
		c <- struct{}{}
	}
}

func (s *S) onceTimerHandler(fn func(), t time.Time, timerID uint) {
	now := time.Now()

	if t.After(now) {
		ticker := time.NewTicker(t.Sub(now))
		defer ticker.Stop()
		s.timerIDMutex.RLock()
		c := s.timerIDs[timerID]
		s.timerIDMutex.RUnlock()
		select {
		case <-ticker.C:
			s.handlerChan <- fn
		case <-c:
		}
	} else {
		s.handlerChan <- fn
	}

	s.timerIDMutex.Lock()
	defer s.timerIDMutex.Unlock()
	delete(s.timerIDs, timerID)
}

func (s *S) loopTimerHandler(fn func(), now time.Time, d time.Duration, timerID uint, immediately bool) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	if immediately {
		s.handlerChan <- fn
	}

	s.timerIDMutex.RLock()
	c := s.timerIDs[timerID]
	s.timerIDMutex.RUnlock()

FOR_LOOP:
	for {
		select {
		case <-ticker.C:
			s.handlerChan <- fn
		case <-c:
			break FOR_LOOP
		}
	}
	s.timerIDMutex.Lock()
	defer s.timerIDMutex.Unlock()
	delete(s.timerIDs, timerID)
}

// NumWaiting 等待队列还有多少
func (s *S) NumWaiting() int {
	return len(s.handlerChan)
}

// NumTimer 定时器数量
func (s *S) NumTimer() int {
	return len(s.timerIDs)
}
