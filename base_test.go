package sync

import (
	"strconv"
	"sync"
	"testing"
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

func Test_NoDeadlock(t *testing.T) {
	var (
		mutex Mutex
		wait  WaitGroup
	)

	wait.Add(1)
	go func() {
		for i := 0; i < 10000; i++ {
			mutex.Lock()
			strconv.Itoa(i)
			mutex.Unlock()
		}
		wait.Done()
	}()

	wait.Add(1)
	go func() {
		for i := 0; i < 10000; i++ {
			mutex.Lock()
			strconv.Itoa(i)
			mutex.Unlock()
		}
		wait.Done()
	}()

	wait.Wait()
}
