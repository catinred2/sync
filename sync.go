package sync

import (
	"bytes"
	"container/list"
	"fmt"
	"github.com/funny/goid"
	"runtime/pprof"
	"strconv"
	"sync"
	"sync/atomic"
)

var WatchDeadLock int32 = 1

var (
	lockMutex      = new(sync.Mutex)
	waitTargets    = make(map[int32]*mutexInfo)
	goroutineBegin = []byte("goroutine ")
	goroutineEnd   = []byte("\n\n")
)

func goroutine(id int32, stack []byte) []byte {
	head := append(strconv.AppendInt(goroutineBegin, int64(id), 10), ' ')
	begin := bytes.Index(stack, head)
	end := bytes.Index(stack[begin:], goroutineEnd)
	if end == -1 {
		end = len(stack) - begin
	}
	return stack[begin : begin+end]
}

type mutexInfo struct {
	holder  int32
	waiting *list.List
	watch   bool
}

func (lock *mutexInfo) wait() (int32, *list.Element) {
	lockMutex.Lock()
	defer lockMutex.Unlock()

	holder := goid.Get()
	waitTargets[holder] = lock

	if lock.waiting == nil {
		lock.waiting = list.New()
	}

	lock.verify(holder, []*mutexInfo{lock})
	return holder, lock.waiting.PushBack(holder)
}

func (lock *mutexInfo) verify(holder int32, link []*mutexInfo) {
	if lock.holder != 0 {
		if lock.holder == holder {
			stackBuf := new(bytes.Buffer)
			prof := pprof.Lookup("goroutine")
			prof.WriteTo(stackBuf, 2)
			stack := stackBuf.Bytes()

			buf := new(bytes.Buffer)
			fmt.Fprintln(buf, "[DEAD LOCK]\n")
			fmt.Fprintf(buf, "%s\n\n", goroutine(holder, stack))
			for i := 0; i < len(link); i++ {
				fmt.Fprintf(buf, "%s\n\n", goroutine(link[i].holder, stack))
			}
			panic(buf.String())
		}
		if waitTarget, exists := waitTargets[lock.holder]; exists {
			waitTarget.verify(holder, append(link, waitTarget))
		}
	}
}

func (lock *mutexInfo) using(holder int32, elem *list.Element) {
	lockMutex.Lock()
	defer lockMutex.Unlock()

	delete(waitTargets, holder)
	atomic.StoreInt32(&lock.holder, holder)
	lock.waiting.Remove(elem)
}

func (lock *mutexInfo) release() {
	atomic.StoreInt32(&lock.holder, 0)
}

type Mutex struct {
	mutexInfo
	sync.Mutex
}

func (m *Mutex) Lock() {
	m.mutexInfo.watch = atomic.LoadInt32(&WatchDeadLock) == 1

	if m.mutexInfo.watch {
		holder, elem := m.mutexInfo.wait()
		m.Mutex.Lock()
		m.mutexInfo.using(holder, elem)
	} else {
		m.Mutex.Lock()
	}
}

func (m *Mutex) Unlock() {
	if m.mutexInfo.watch {
		m.mutexInfo.release()
	}
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
