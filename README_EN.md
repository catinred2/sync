Introduction
============

This package is used to detect deadlock in Go program.

Usage
=====

Normally, we import default `sync` package in our project like this：

```go
package myapp

import "sync"

var MyLock sync.Mutex

func MyFunc() {
	MyLock.Lock()
	defer MyLock.Unlock()

	// .......
}
```

Just replace the default `sync` to `github.com/funny/sync`, no need to change others：


```go
package myapp

import "github.com/funny/sync"

var MyLock sync.Mutex

func MyFunc() {
	MyLock.Lock()
	defer MyLock.Unlock()

	// .......
}
```

Currently, deadlock detection not yet enabled, the performance of `Mutext` and `RWMutex` just like default.

When you need to compile a deadlock detection enabled version. Just add `deadlock` tag into `go build --tags` command.

For example:

```
go build -tags 'deadlock' myproject
```

This tag used for the unit test too. Otherwise the default unit test will deadlock:

```
go test -tags 'deadlock'
```

How it works
============

When deadlock detection enabled, system will maintain a global lock waiting list, and each `Mutex` and `RWMutex` will keep owner goroutine's information.

When a goroutine will waiting a lock, system will lookup the lock owner goroutine is whether waiting for the requester goroutine in directly or indirectly.

Deadlock not only happens between two goroutines, sometimes the deadlock is a link, and deadlock happens when a goroutine repeat lock a `Mutext` too.

When deadlock happens, system will dump stack trace of the gorotuines in the deadlock link.

Because we need a global lock waiting list, so the deadlock detection will drop the performance.

So, please don't use deadlock detection in production environment.