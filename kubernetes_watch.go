package lambda

import (
	"reflect"
	"sync"

	"k8s.io/apimachinery/pkg/watch"
)

type watchStopCh chan struct{}
type watchFunction struct {
	function Function
	t        watch.EventType
}

type watchEntry struct {
	stopCh         watchStopCh
	watchFunctions []watchFunction
}

func (we *watchEntry) AddFunc(t watch.EventType, f Function) {
	we.watchFunctions = append(we.watchFunctions, watchFunction{
		function: f,
		t:        t,
	})
}

func (we *watchEntry) DelFunc(t watch.EventType, f Function) {
	for index, wf := range we.watchFunctions {
		if reflect.ValueOf(wf.function).Pointer() == reflect.ValueOf(f).Pointer() && wf.t == t {
			we.watchFunctions = append(we.watchFunctions[:index], we.watchFunctions[index+1:]...)
		}
	}
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
		watchManagerInstance = &watchManager{
			watchStopChs: make(resourcedEntries),
		}
	})
	return watchManagerInstance
}

func (wm *watchManager) registerFunc(rs Resource, ns string, t watch.EventType, function Function) *watchEntry {
	wm.rwlock.Lock()
	defer wm.rwlock.Unlock()
	wm.preStop(rs, ns)
	entry := wm.getEntry(rs, ns)
	entry.AddFunc(t, function)
	return entry
}

func (wm *watchManager) unregisterFunc(rs Resource, ns string, t watch.EventType, function Function) *watchEntry {
	wm.rwlock.Lock()
	defer wm.rwlock.Unlock()
	wm.preStop(rs, ns)
	entry := wm.getEntry(rs, ns)
	entry.DelFunc(t, function)
	return entry
}

func (wm *watchManager) preStop(rs Resource, ns string) {
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
		stopCh:         ch,
		watchFunctions: []watchFunction{},
	}
}

func (wm *watchManager) getEntry(rs Resource, ns string) *watchEntry {
	if _, exists := wm.watchStopChs[rs.String()]; !exists {
		wm.watchStopChs[rs.String()] = make(namespacedEntries)
	}
	return wm.watchStopChs[rs.String()][ns]
}
