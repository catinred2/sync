package sync

import (
	"fmt"
	"github.com/funny/debug"
	"github.com/funny/goid"
	"os"
	"strconv"
	"sync"
	"time"
)

var _GO_LOCK_TIMEOUT_ time.Duration

func init() {
	if t, e := strconv.Atoi(debug.GODEBUG("locktimeout", "10")); e != nil {
		_GO_LOCK_TIMEOUT_ = time.Second * time.Duration(t)
	}
}

func LockTimeout(n int) (r int) {
	r = int(_GO_LOCK_TIMEOUT_ / time.Second)
	if n != 0 {
		_GO_LOCK_TIMEOUT_ = time.Second * time.Duration(n)
	}
	return
}

type lockDebugger struct {
	lockBy       int32
	timeout      time.Duration
	timeoutPanic bool
	timer        *time.Timer
}

func (debugger *lockDebugger) checkOwner() {
	id := goid.Get()
	if debugger.lockBy != 0 && debugger.lockBy == id {
		stack := debug.StackTrace(3, 0)
		panic(fmt.Sprintf("PANIC: DUPLICATE LOCK\nLock by goroutine %d\n%s\n", id, stack.Bytes("  ")))
	}
}

func (debugger *lockDebugger) start() {
	id := goid.Get()
	debugger.lockBy = id

	if _GO_LOCK_TIMEOUT_ > 0 || debugger.timeout > 0 {
		timeout := _GO_LOCK_TIMEOUT_
		if debugger.timeout != 0 {
			timeout = debugger.timeout
		}
		stack := debug.StackTrace(3, 0)
		debugger.timer = time.AfterFunc(timeout, func() {
			if debugger.timeoutPanic {
				panic(fmt.Sprintf("PANIC: LOCK TIMEOUT\nLock by goroutine %d\n%s\n", id, stack.Bytes("  ")))
			} else {
				fmt.Fprintf(os.Stderr, "WARNING: LOCK TIMEOUT\nLock by goroutine %d\n%s\n", id, stack.Bytes("  "))
			}
		})
	}
}

func (debugger *lockDebugger) stop() {
	debugger.lockBy = 0
	if debugger.timer != nil {
		debugger.timer.Stop()
	}
	debugger.timer = nil
}

func (debugger *lockDebugger) LockTimeout(timeout int) {
	debugger.timeout = time.Second * time.Duration(timeout)
}

func (debugger *lockDebugger) TimeoutPanic(enable bool) {
	debugger.timeoutPanic = enable
}

type Mutex struct {
	lockDebugger
	m sync.Mutex
}

func (m *Mutex) Lock() {
	m.checkOwner()
	m.m.Lock()
	m.start()
}

func (m *Mutex) Unlock() {
	m.stop()
	m.m.Unlock()
}

type RWMutex struct {
	lockDebugger
	m sync.RWMutex
}

func (rw *RWMutex) Lock() {
	rw.checkOwner()
	rw.m.Lock()
	rw.start()
}

func (rw *RWMutex) Unlock() {
	rw.stop()
	rw.m.Unlock()
}

func (rw *RWMutex) RLock() {
	rw.m.RLock()
}

func (rw *RWMutex) RUnlock() {
	rw.m.RUnlock()
}
