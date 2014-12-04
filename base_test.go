package sync

import (
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
