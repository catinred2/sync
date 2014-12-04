// +build deadlock

package sync

import (
	"testing"
	"time"
)

func init() {
	go func() {
		for {
			time.Sleep(time.Second)
		}
	}()
}

func testRecover(err interface{}, testDone chan int) {
	if err != nil {
		switch err.(type) {
		case DeadlockError:
			if testing.Verbose() {
				println(err.(DeadlockError).Error())
			}
			testDone <- 1
		default:
			panic(err)
		}
	}
}

func Test_DeadLock1(t *testing.T) {
	testDone := make(chan int)

	var mutex1 Mutex

	go func() {
		defer func() {
			err := recover()
			testRecover(err, testDone)
		}()

		mutex1.Lock()
		mutex1.Lock()
	}()

	select {
	case <-testDone:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func Test_DeadLock2(t *testing.T) {
	testDone := make(chan int)

	var (
		mutex1 Mutex
		mutex2 Mutex

		wait1 WaitGroup
	)

	wait1.Add(1)
	go func() {
		defer func() {
			err := recover()
			testRecover(err, testDone)
		}()

		mutex1.Lock()
		wait1.Wait()
		mutex2.Lock()
	}()

	go func() {
		mutex2.Lock()
		wait1.Done()
		mutex1.Lock()
	}()

	select {
	case <-testDone:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func Test_DeadLock3(t *testing.T) {
	testDone := make(chan int)

	var (
		mutex1 Mutex
		mutex2 Mutex
		mutex3 Mutex

		wait1 WaitGroup
		wait2 WaitGroup
	)

	wait1.Add(1)
	wait2.Add(1)

	go func() {
		defer func() {
			err := recover()
			testRecover(err, testDone)
		}()

		mutex1.Lock()
		wait1.Wait()
		mutex2.Lock()
	}()

	go func() {
		mutex2.Lock()
		wait2.Wait()
		wait1.Done()
		mutex3.Lock()
	}()

	go func() {
		mutex3.Lock()
		wait2.Done()
		mutex1.Lock()
	}()

	select {
	case <-testDone:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
