package lambda

import (
	"sync"
)

type watchStopCh chan struct{}

type watchEntry struct {
	stopCh    watchStopCh
	functions []Function
}

func (we *watchEntry) AddFunc(f Function) {
	we.functions = append(we.functions, f)
}

type namespacedEntries map[string]*watchEntry

type resourcedEntries map[string]namespacedEntries

var (
	watchManagerInstance *watchManager
	once                 sync.Once
)

type watchManager struct {
	rwlock       sync.RWMutex
	watchStopChs resourcedEntries
}

func getWatchManager() *watchManager {
	once.Do(func() {
		//instance := &watchManager{
		//	watchStopChs: make(resourcedEntries),
		//}
	})
	return nil
}

func (wm *watchManager) register(rs Resource, ns string, function Function) {
	wm.preStop(rs, ns)
	entry := wm.getEntry(rs, ns)
	entry.AddFunc(function)

}

func (wm *watchManager) preStop(rs Resource, ns string) {
	wm.rwlock.Lock()
	defer wm.rwlock.Unlock()
	if _, exists := wm.watchStopChs[rs.String()]; !exists {
		wm.watchStopChs[rs.String()] = make(namespacedEntries)
	}
	ch := make(chan struct{})
	if e, exists := wm.watchStopChs[rs.String()][ns]; exists {
		e.stopCh <- struct{}{}
		close(e.stopCh)
		e.stopCh = ch
		return
	}
	wm.watchStopChs[rs.String()][ns] = &watchEntry{
		stopCh:    ch,
		functions: []Function{},
	}
}

func (wm *watchManager) start(entry *watchEntry) {
	go func() {

	}()

}

func (wm *watchManager) getEntry(rs Resource, ns string) *watchEntry {
	if _, exists := wm.watchStopChs[rs.String()]; !exists {
		wm.watchStopChs[rs.String()] = make(namespacedEntries)
	}
	return wm.watchStopChs[rs.String()][ns]
}
