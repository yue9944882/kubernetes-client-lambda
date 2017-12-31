package lambda

import (
	"fmt"
	"reflect"
	"sync"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/testing"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

const (
	watchEventMaxBufSize = 1024
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
	ch := make(chan struct{}, watchEventMaxBufSize)
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

type kubernetesObjectTracker interface {
	testing.ObjectTracker
	// Watch watches objects from the tracker. Watch returns a channel
	// which will push added / modified / deleted object.
	Watch(gvr schema.GroupVersionResource, ns string) (watch.Interface, error)
}

var _ kubernetesObjectTracker = &kubernetesTracker{}

type kubernetesTracker struct {
	scheme   testing.ObjectScheme
	decoder  runtime.Decoder
	lock     sync.RWMutex
	objects  map[schema.GroupVersionResource][]runtime.Object
	watchers map[schema.GroupVersionResource]map[string]*watch.FakeWatcher
}

func (t *kubernetesTracker) List(gvr schema.GroupVersionResource, gvk schema.GroupVersionKind, ns string) (runtime.Object, error) {
	// Heuristic for list kind: original kind + List suffix. Might
	// not always be true but this tracker has a pretty limited
	// understanding of the actual API model.
	listGVK := gvk
	listGVK.Kind = listGVK.Kind + "List"
	// GVK does have the concept of "internal version". The scheme recognizes
	// the runtime.APIVersionInternal, but not the empty string.
	if listGVK.Version == "" {
		listGVK.Version = runtime.APIVersionInternal
	}

	list, err := t.scheme.New(listGVK)
	if err != nil {
		return nil, err
	}

	if !meta.IsListType(list) {
		return nil, fmt.Errorf("%q is not a list type", listGVK.Kind)
	}

	t.lock.RLock()
	defer t.lock.RUnlock()

	objs, ok := t.objects[gvr]
	if !ok {
		return list, nil
	}

	matchingObjs, err := filterByNamespaceAndName(objs, ns, "")
	if err != nil {
		return nil, err
	}
	if err := meta.SetList(list, matchingObjs); err != nil {
		return nil, err
	}
	return list.DeepCopyObject(), nil
}

func (t *kubernetesTracker) Watch(gvr schema.GroupVersionResource, ns string) (watch.Interface, error) {
	fakewatcher := watch.NewFake()
	if _, exists := t.watchers[gvr]; !exists {
		t.watchers[gvr] = make(map[string]*watch.FakeWatcher)
	}
	if _, exists := t.watchers[gvr][ns]; !exists {
		t.watchers[gvr][ns] = fakewatcher
		return fakewatcher, nil
	}
	return t.watchers[gvr][ns], nil
}

func (t *kubernetesTracker) Get(gvr schema.GroupVersionResource, ns, name string) (runtime.Object, error) {
	errNotFound := errors.NewNotFound(gvr.GroupResource(), name)

	t.lock.RLock()
	defer t.lock.RUnlock()

	objs, ok := t.objects[gvr]
	if !ok {
		return nil, errNotFound
	}

	matchingObjs, err := filterByNamespaceAndName(objs, ns, name)
	if err != nil {
		return nil, err
	}
	if len(matchingObjs) == 0 {
		return nil, errNotFound
	}
	if len(matchingObjs) > 1 {
		return nil, fmt.Errorf("more than one object matched gvr %s, ns: %q name: %q", gvr, ns, name)
	}

	// Only one object should match in the tracker if it works
	// correctly, as Add/Update methods enforce kind/namespace/name
	// uniqueness.
	obj := matchingObjs[0].DeepCopyObject()
	if status, ok := obj.(*meta_v1.Status); ok {
		if status.Status != meta_v1.StatusSuccess {
			return nil, &errors.StatusError{ErrStatus: *status}
		}
	}

	return obj, nil
}

func (t *kubernetesTracker) add(gvr schema.GroupVersionResource, obj runtime.Object, ns string, replaceExisting bool) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	gr := gvr.GroupResource()

	// To avoid the object from being accidentally modified by caller
	// after it's been added to the tracker, we always store the deep
	// copy.
	obj = obj.DeepCopyObject()

	newMeta, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	// Propagate namespace to the new object if hasn't already been set.
	if len(newMeta.GetNamespace()) == 0 {
		newMeta.SetNamespace(ns)
	}

	if ns != newMeta.GetNamespace() {
		msg := fmt.Sprintf("request namespace does not match object namespace, request: %q object: %q", ns, newMeta.GetNamespace())
		return errors.NewBadRequest(msg)
	}

	for i, existingObj := range t.objects[gvr] {
		oldMeta, err := meta.Accessor(existingObj)
		if err != nil {
			return err
		}
		if oldMeta.GetNamespace() == newMeta.GetNamespace() && oldMeta.GetName() == newMeta.GetName() {
			if replaceExisting {
				t.objects[gvr][i] = obj
				return nil
			}
			return errors.NewAlreadyExists(gr, newMeta.GetName())
		}
	}

	if replaceExisting {
		// Tried to update but no matching object was found.
		return errors.NewNotFound(gr, newMeta.GetName())
	}

	t.objects[gvr] = append(t.objects[gvr], obj)

	return nil
}

func (t *kubernetesTracker) Add(obj runtime.Object) error {
	if meta.IsListType(obj) {
		return t.addList(obj, false)
	}
	objMeta, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	gvks, _, err := t.scheme.ObjectKinds(obj)
	if err != nil {
		return err
	}
	if len(gvks) == 0 {
		return fmt.Errorf("no registered kinds for %v", obj)
	}
	for _, gvk := range gvks {
		// NOTE: UnsafeGuessKindToResource is a heuristic and default match. The
		// actual registration in apiserver can specify arbitrary route for a
		// gvk. If a test uses such objects, it cannot preset the tracker with
		// objects via Add(). Instead, it should trigger the Create() function
		// of the tracker, where an arbitrary gvr can be specified.
		gvr, _ := meta.UnsafeGuessKindToResource(gvk)
		// Resource doesn't have the concept of "__internal" version, just set it to "".
		if gvr.Version == runtime.APIVersionInternal {
			gvr.Version = ""
		}

		err := t.add(gvr, obj, objMeta.GetNamespace(), false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *kubernetesTracker) Create(gvr schema.GroupVersionResource, obj runtime.Object, ns string) error {
	err := t.add(gvr, obj, ns, false)
	if err != nil {
		return err
	}
	if t.watchers[gvr] != nil && t.watchers[gvr][ns] != nil {
		t.watchers[gvr][ns].Add(obj)
	}
	return nil
}

func (t *kubernetesTracker) Update(gvr schema.GroupVersionResource, obj runtime.Object, ns string) error {
	err := t.add(gvr, obj, ns, true)
	if err != nil {
		return err
	}
	if t.watchers[gvr] != nil && t.watchers[gvr][ns] != nil {
		t.watchers[gvr][ns].Modify(obj)
	}
	return nil
}

func (t *kubernetesTracker) Delete(gvr schema.GroupVersionResource, ns, name string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	found := false

	for i, existingObj := range t.objects[gvr] {
		objMeta, err := meta.Accessor(existingObj)
		if err != nil {
			return err
		}
		if objMeta.GetNamespace() == ns && objMeta.GetName() == name {
			obj := t.objects[gvr][i]
			t.objects[gvr] = append(t.objects[gvr][:i], t.objects[gvr][i+1:]...)
			if t.watchers[gvr] != nil && t.watchers[gvr][ns] != nil {
				t.watchers[gvr][ns].Delete(obj)
			}
			found = true
			break
		}
	}

	if found {
		return nil
	}

	return errors.NewNotFound(gvr.GroupResource(), name)
}

func filterByNamespaceAndName(objs []runtime.Object, ns, name string) ([]runtime.Object, error) {
	var res []runtime.Object

	for _, obj := range objs {
		acc, err := meta.Accessor(obj)
		if err != nil {
			return nil, err
		}
		if ns != "" && acc.GetNamespace() != ns {
			continue
		}
		if name != "" && acc.GetName() != name {
			continue
		}
		res = append(res, obj)
	}

	return res, nil
}

func (t *kubernetesTracker) addList(obj runtime.Object, replaceExisting bool) error {
	list, err := meta.ExtractList(obj)
	if err != nil {
		return err
	}
	errs := runtime.DecodeList(list, t.decoder)
	if len(errs) > 0 {
		return errs[0]
	}
	for _, obj := range list {
		if err := t.Add(obj); err != nil {
			return err
		}
	}
	return nil
}
