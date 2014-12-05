// +build deadlock

package sync

import (
	"bytes"
	"container/list"
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
	waitInfo := m.monitor.wait('w')
	m.Mutex.Lock()
	m.monitor.using(waitInfo)
}

func (m *Mutex) Unlock() {
	m.monitor.release('w')
	m.Mutex.Unlock()
}

type RWMutex struct {
	monitor
	sync.RWMutex
}

func (m *RWMutex) Lock() {
	waitInfo := m.monitor.wait('w')
	m.RWMutex.Lock()
	m.monitor.using(waitInfo)
}

func (m *RWMutex) Unlock() {
	m.monitor.release('w')
	m.RWMutex.Unlock()
}

func (m *RWMutex) RLock() {
	waitInfo := m.monitor.wait('r')
	m.RWMutex.RLock()
	m.monitor.using(waitInfo)
}

func (m *RWMutex) RUnlock() {
	m.monitor.release('r')
	m.RWMutex.RUnlock()
}

var (
	globalMutex = new(sync.Mutex)
	waitingList = make(map[int32]*waiting)
	titleStr    = []byte("[DEAD LOCK]\n")
	goStr       = []byte("goroutine ")
	waitStr     = []byte(" wait")
	holdStr     = []byte(" hold")
	readStr     = []byte(" read")
	writeStr    = []byte(" write")
	lineStr     = []byte{'\n'}
)

type waiting struct {
	monitor     *monitor
	mode        byte
	holder      int32
	holderStack debug.StackInfo
}

type monitor struct {
	holders *list.List
}

func (m *monitor) wait(mode byte) *waiting {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	waitInfo := &waiting{m, mode, goid.Get(), debug.StackTrace(3, 0)}
	waitingList[waitInfo.holder] = waitInfo

	if m.holders == nil {
		m.holders = list.New()
	}

	m.verify(mode, []*waiting{waitInfo})

	return waitInfo
}

func (m *monitor) verify(mode byte, waitLink []*waiting) {
	for i := m.holders.Front(); i != nil; i = i.Next() {
		holder := i.Value.(*waiting)
		if mode != 'r' || holder.mode != 'r' {
			// deadlock detected
			if holder.holder == waitLink[0].holder {
				buf := new(bytes.Buffer)
				buf.Write(titleStr)
				for i := 0; i < len(waitLink); i++ {
					buf.Write(goStr)
					buf.WriteString(strconv.Itoa(int(waitLink[i].holder)))
					buf.Write(waitStr)
					if waitLink[i].mode == 'w' {
						buf.Write(writeStr)
					} else {
						buf.Write(readStr)
					}
					buf.Write(lineStr)
					buf.Write(waitLink[i].holderStack.Bytes("  "))

					// lookup waiting for who
					n := i + 1
					if n == len(waitLink) {
						n = 0
					}
					waitWho := waitLink[n]

					for j := waitLink[i].monitor.holders.Front(); j != nil; j = j.Next() {
						waitHolder := j.Value.(*waiting)
						if waitHolder.holder == waitWho.holder {
							buf.Write(goStr)
							buf.WriteString(strconv.Itoa(int(waitHolder.holder)))
							buf.Write(holdStr)
							if waitHolder.mode == 'w' {
								buf.Write(writeStr)
							} else {
								buf.Write(readStr)
							}
							buf.Write(lineStr)
							buf.Write(waitHolder.holderStack.Bytes("  "))
							break
						}
					}
				}
				panic(DeadlockError(buf.String()))
			}
			// the lock holder is waiting for another lock
			if waitInfo, exists := waitingList[holder.holder]; exists {
				waitInfo.monitor.verify(waitInfo.mode, append(waitLink, waitInfo))
			}
		}
	}
}

func (m *monitor) using(waitInfo *waiting) {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	delete(waitingList, waitInfo.holder)
	m.holders.PushBack(waitInfo)
}

func (m *monitor) release(mode byte) {
	holder := goid.Get()
	for i := m.holders.Back(); i != nil; i = i.Prev() {
		if info := i.Value.(*waiting); info.holder == holder && info.mode == mode {
			m.holders.Remove(i)
			break
		}
	}
}
