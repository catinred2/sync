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
	monitor
	sync.Mutex
}

func (m *Mutex) Lock() {
	waitInfo := m.monitor.wait()
	m.Mutex.Lock()
	m.monitor.using(waitInfo)
}

func (m *Mutex) Unlock() {
	m.monitor.release()
	m.Mutex.Unlock()
}

type RWMutex struct {
	monitor
	sync.RWMutex
}

func (m *RWMutex) Lock() {
	waitInfo := m.monitor.wait()
	m.RWMutex.Lock()
	m.monitor.using(waitInfo)
}

func (m *RWMutex) Unlock() {
	m.monitor.release()
	m.RWMutex.Unlock()
}

func (m *RWMutex) RLock() {
	m.RWMutex.RLock()
}

func (m *RWMutex) RUnlock() {
	m.RWMutex.RUnlock()
}

var (
	globalMutex = new(sync.Mutex)
	waitingList = make(map[int32]*waiting)
	titleStr    = []byte("[DEAD LOCK]\n")
	goStr       = []byte("goroutine ")
	waitStr     = []byte(" wait")
	holdStr     = []byte(" hold")
	lineStr     = []byte{'\n'}
)

type monitor struct {
	holder      int32
	holderStack debug.StackInfo
}

type waiting struct {
	monitor     *monitor
	holder      int32
	holderStack debug.StackInfo
}

func (m *monitor) wait() *waiting {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	waitInfo := &waiting{m, goid.Get(), debug.StackTrace(3, 0)}
	waitingList[waitInfo.holder] = waitInfo

	m.verify([]*waiting{waitInfo})

	return waitInfo
}

func (m *monitor) verify(waitLink []*waiting) {
	if m.holder != 0 {
		// deadlock detected
		if m.holder == waitLink[0].holder {
			buf := new(bytes.Buffer)
			buf.Write(titleStr)
			for i := 0; i < len(waitLink); i++ {
				buf.Write(goStr)
				buf.WriteString(strconv.Itoa(int(waitLink[i].holder)))
				buf.Write(waitStr)
				buf.Write(lineStr)
				buf.Write(waitLink[i].holderStack.Bytes("  "))

				buf.Write(goStr)
				buf.WriteString(strconv.Itoa(int(waitLink[i].monitor.holder)))
				buf.Write(holdStr)
				buf.Write(lineStr)
				buf.Write(waitLink[i].monitor.holderStack.Bytes("  "))
			}
			panic(DeadlockError(buf.String()))
		}
		// the lock holder is waiting for another lock
		if waitInfo, exists := waitingList[m.holder]; exists {
			waitInfo.monitor.verify(append(waitLink, waitInfo))
		}
	}
}

func (m *monitor) using(waitInfo *waiting) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	delete(waitingList, waitInfo.holder)
	m.holder = waitInfo.holder
	m.holderStack = waitInfo.holderStack
}

func (m *monitor) release() {
	m.holder = 0
	m.holderStack = nil
}
