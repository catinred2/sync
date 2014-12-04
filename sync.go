// +build deadlock

package sync

import (
	"bytes"
	"fmt"
	"github.com/funny/goid"
	"runtime/pprof"
	"strconv"
	"sync"
	"sync/atomic"
)

type Mutex struct {
	mointor
	sync.Mutex
}

func (m *Mutex) Lock() {
	holder := m.mointor.wait()
	m.Mutex.Lock()
	m.mointor.using(holder)
}

func (m *Mutex) Unlock() {
	m.mointor.release()
	m.Mutex.Unlock()
}

type RWMutex struct {
	Mutex
}

func (rw *RWMutex) Lock() {
	rw.Lock()
}

func (rw *RWMutex) Unlock() {
	rw.Unlock()
}

func (rw *RWMutex) RLock() {
	rw.Lock()
}

func (rw *RWMutex) RUnlock() {
	rw.Unlock()
}

var (
	globalMutex    = new(sync.Mutex)
	waitTargets    = make(map[int32]*mointor)
	goroutineBegin = []byte("goroutine ")
	goroutineEnd   = []byte("\n\n")
)

type mointor struct {
	holder int32
}

func (m *mointor) wait() int32 {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	holder := goid.Get()
	waitTargets[holder] = m

	m.verify(holder, []int32{holder})
	return holder
}

func (m *mointor) verify(holder int32, holderLink []int32) {
	if m.holder != 0 {
		if m.holder == holder {
			buf := new(bytes.Buffer)
			fmt.Fprintln(buf, "[DEAD LOCK]\n")

			// dump goroutines
			stackBuf := new(bytes.Buffer)
			prof := pprof.Lookup("goroutine")
			prof.WriteTo(stackBuf, 2)
			stack := stackBuf.Bytes()
			fmt.Fprintf(buf, "%s\n\n", traceGoroutine(holder, stack))
			for i := 0; i < len(holderLink); i++ {
				fmt.Fprintf(buf, "%s\n\n", traceGoroutine(holderLink[i], stack))
			}

			panic(buf.String())
		}
		if waitTarget, exists := waitTargets[m.holder]; exists {
			waitTarget.verify(holder, append(holderLink, waitTarget.holder))
		}
	}
}

func (m *mointor) using(holder int32) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	delete(waitTargets, holder)
	atomic.StoreInt32(&m.holder, holder)
}

func (m *mointor) release() {
	atomic.StoreInt32(&m.holder, 0)
}

func traceGoroutine(id int32, stack []byte) []byte {
	head := append(strconv.AppendInt(goroutineBegin, int64(id), 10), ' ')
	begin := bytes.Index(stack, head)
	end := bytes.Index(stack[begin:], goroutineEnd)
	if end == -1 {
		end = len(stack) - begin
	}
	return stack[begin : begin+end]
}
