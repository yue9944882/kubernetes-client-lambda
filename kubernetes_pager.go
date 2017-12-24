package lambda

import (
	"reflect"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ListPageFunc interface{}

type ListPager struct {
	PageSize int64
	PageFn   ListPageFunc
}

func (p *ListPager) List(options meta_v1.ListOptions) (<-chan kubernetesResource, error) {
	if options.Limit == 0 {
		options.Limit = p.PageSize
	}
	ch := make(chan kubernetesResource)
	go func() {
		defer close(ch)
		for {
			method := reflect.ValueOf(p.PageFn)
			ret := method.Call([]reflect.Value{
				reflect.ValueOf(options),
			})
			if err := ret[1].Interface(); err != nil {
				panic(err.(error))
			}
			obj := ret[0].Interface().(runtime.Object)
			m, err := meta.ListAccessor(obj)
			if err != nil {
				panic(err)
			}
			items, err := meta.ExtractList(obj)
			if err != nil {
				panic(err)
			}

			count := 0
			for _, item := range items {
				ch <- item
				count++
			}

			// if we have no more items, return the list
			if len(m.GetContinue()) == 0 {
				return
			}
			// set the next loop up
			options.Continue = m.GetContinue()
		}
	}()
	return ch, nil
}
