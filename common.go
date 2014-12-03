package sync

import "sync"

type Cond sync.Cond

func NewCond(l Locker) *Cond {
	return (*Cond)(sync.NewCond(l))
}

type Locker sync.Locker
type Once sync.Once
type Pool sync.Pool
type WaitGroup sync.WaitGroup
