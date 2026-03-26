package handlers

import (
	"sync"
)

var (
	scanLocks = make(map[string]*sync.Mutex)
	locksMu   sync.Mutex
)

func getScanLock(id string) *sync.Mutex {
	locksMu.Lock()
	defer locksMu.Unlock()
	if _, ok := scanLocks[id]; !ok {
		scanLocks[id] = &sync.Mutex{}
	}
	return scanLocks[id]
}

// TryLockScan attempts to acquire a lock for a scan ID. Returns true if acquired.
func TryLockScan(id string) bool {
	mu := getScanLock(id)
	return mu.TryLock()
}

// UnlockScan releases the lock for a scan ID.
func UnlockScan(id string) {
	mu := getScanLock(id)
	mu.Unlock()
}
