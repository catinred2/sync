// +build deadlock

package sync

import (
	"bytes"
	"github.com/funny/debug"
	"github.com/funny/goid"
	"strconv"
	"sync"
)

type Mutex struct {
	mointor
	sync.Mutex
}

func (m *Mutex) Lock() {
	holder, holderStack := m.mointor.wait()
	m.Mutex.Lock()
	m.mointor.using(holder, holderStack)
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
	newline        = []byte{'\n'}
)

type mointor struct {
	holder      int32
	holderStack debug.StackInfo
}

func (m *mointor) wait() (int32, debug.StackInfo) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	holder := goid.Get()
	holderStack := debug.StackTrace(3, 0)
	waitTargets[holder] = m

	m.verify([]*mointor{{holder, holderStack}})

	return holder, holderStack
}

func (m *mointor) verify(holderLink []*mointor) {
	if m.holder != 0 {
		// deadlock detected
		if m.holder == holderLink[0].holder {
			buf := new(bytes.Buffer)
			buf.WriteString("[DEAD LOCK]\n")
			for i := 0; i < len(holderLink); i++ {
				buf.Write(goroutineBegin)
				buf.WriteString(strconv.Itoa(int(holderLink[i].holder)))
				buf.Write(newline)
				buf.Write(holderLink[i].holderStack.Bytes("  "))
			}
			panic(DeadlockError(buf.String()))
		}
		// the lock holder is waiting for another lock
		if waitTarget, exists := waitTargets[m.holder]; exists {
			waitTarget.verify(append(holderLink, m))
		}
	}
}

func (m *mointor) using(holder int32, holderStack debug.StackInfo) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	delete(waitTargets, holder)
	m.holder = holder
	m.holderStack = holderStack
}

func (m *mointor) release() {
	m.holder = 0
	m.holderStack = nil
}
