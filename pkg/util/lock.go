package util

import (
	"sync"
)

var applock *sync.RWMutex = new(sync.RWMutex)
var apilock *sync.RWMutex = new(sync.RWMutex)
var sulock *sync.RWMutex = new(sync.RWMutex)

func ApplicationRLock() {
	applock.RLock()
}

func ApplicationRUnlock() {
	applock.RUnlock()
}

func ApiRLock() {
	apilock.RLock()
}

func ApiRUnlock() {
	apilock.RUnlock()
}

func ServiceunitRLock() {
	sulock.RLock()
}

func ServiceunitRUnlock() {
	sulock.RUnlock()
}
