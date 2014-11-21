package sync

import (
	"testing"
	"time"
)

func Test_LockTimeout(t *testing.T) {
	var mutex Mutex

	LockTimeout(1)
	mutex.Lock()
	time.Sleep(time.Second * 2)
	mutex.Unlock()

	LockTimeout(3)
	mutex.Lock()
	time.Sleep(time.Second * 2)
	mutex.Unlock()

	mutex.LockTimeout(1)
	mutex.Lock()
	time.Sleep(time.Second * 2)
	mutex.Unlock()
}

func Test_DuplicateLock(t *testing.T) {
	err := func() (e interface{}) {
		defer func() {
			e = recover()
		}()
		var mutex Mutex
		mutex.Lock()
		mutex.Lock()
		return
	}()
	if err == nil {
		t.Fatal("duplicate lock not panic")
	}

	print(err.(string))
}
