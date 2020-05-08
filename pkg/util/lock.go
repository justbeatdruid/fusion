package util

import (
	"sync"
)

var applock *sync.Mutex = new(sync.Mutex)
var apilock *sync.Mutex = new(sync.Mutex)
var sulock *sync.Mutex = new(sync.Mutex)

func ApplicationLock() {
	applock.Lock()
}

func ApplicationUnlock() {
	applock.Unlock()
}

func ApiLock() {
	apilock.Lock()
}

func ApiUnlock() {
	apilock.Unlock()
}

func ServiceunitLock() {
	sulock.Lock()
}

func ServiceunitUnlock() {
	sulock.Unlock()
}
