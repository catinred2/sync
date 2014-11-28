package sync

import (
	"sync"
	"testing"
	"time"
)

func Benchmark_Lock1(b *testing.B) {
	var mutex sync.Mutex
	for i := 0; i < b.N; i++ {
		mutex.Lock()
		mutex.Unlock()
	}
}

func Benchmark_Lock2(b *testing.B) {
	var mutex Mutex
	for i := 0; i < b.N; i++ {
		mutex.Lock()
		mutex.Unlock()
	}
}

func Benchmark_Lock3(b *testing.B) {
	WatchDeadLock = 0
	var mutex Mutex
	for i := 0; i < b.N; i++ {
		mutex.Lock()
		mutex.Unlock()
	}
	WatchDeadLock = 1
}

func Test_DeadLock1(t *testing.T) {
	var testDone sync.WaitGroup
	testDone.Add(1)

	go func() {
		for {
			time.Sleep(time.Second)
		}
	}()

	var mutex1 Mutex

	go func() {
		defer func() {
			println(recover().(string))
			testDone.Done()
		}()
		mutex1.Lock()
		mutex1.Lock()
	}()

	testDone.Wait()
}

func Test_DeadLock2(t *testing.T) {
	var testDone sync.WaitGroup
	testDone.Add(1)

	go func() {
		for {
			time.Sleep(time.Second)
		}
	}()

	var (
		mutex1 Mutex
		mutex2 Mutex

		wait1 sync.WaitGroup
	)

	wait1.Add(1)
	go func() {
		defer func() {
			println(recover().(string))
			testDone.Done()
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

	testDone.Wait()
}

func Test_DeadLock3(t *testing.T) {
	var testDone sync.WaitGroup
	testDone.Add(1)

	go func() {
		for {
			time.Sleep(time.Second)
		}
	}()

	var (
		mutex1 Mutex
		mutex2 Mutex
		mutex3 Mutex

		wait1 sync.WaitGroup
		wait2 sync.WaitGroup
	)

	wait1.Add(1)
	wait2.Add(1)

	go func() {
		defer func() {
			println(recover().(string))
			testDone.Done()
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

	testDone.Wait()
}
