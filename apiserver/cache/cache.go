package cache

import (
	"fmt"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicclient "k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

type gv struct {
	resource string
	gv       schema.GroupVersion
	obj      func() runtime.Object
}

type Store interface {
	Get(namespace, name string) interface{}
	List(namespace string, lo metav1.ListOptions) []interface{}
}

type Listers struct {
	listerMap map[string]cache.GenericLister
	objectMap map[string]func() runtime.Object
}

func (c *Listers) WithResource(r string) (*typedLister, error) {
	lister, ok := c.listerMap[r]
	if !ok {
		return nil, fmt.Errorf("cannot find lister with resource %s", r)
	}
	obj, ok := c.objectMap[r]
	if !ok {
		return nil, fmt.Errorf("cannot find lister with resource %s", r)
	}
	return &typedLister{
		lister:    lister,
		newObject: obj,
	}, nil
}

type typedLister struct {
	lister    cache.GenericLister
	newObject func() runtime.Object
}

func (t *typedLister) Get(namespace, name string) (runtime.Object, error) {
	uobj, err := t.lister.ByNamespace(namespace).Get(name)
	if err != nil {
		return nil, err
	}
	if un, ok := uobj.(*unstructured.Unstructured); ok {
		obj := t.newObject()
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), obj)
		if err != nil {
			return nil, err
		}
		return obj, nil
	}
	return uobj, nil
}

func (t *typedLister) List(namespace string, lo metav1.ListOptions) (uobjs []runtime.Object, err error) {
	selector := labels.NewSelector()
	if len(lo.LabelSelector) > 0 {
		selector, err = labels.Parse(lo.LabelSelector)
		if err != nil {
			return nil, err
		}
	}
	if namespace != metav1.NamespaceAll {
		uobjs, err = t.lister.ByNamespace(namespace).List(selector)
	} else {
		uobjs, err = t.lister.List(selector)
	}
	if err != nil {
		return nil, err
	}
	objs := make([]runtime.Object, len(uobjs))
	for i, uobj := range uobjs {
		if un, ok := uobj.(*unstructured.Unstructured); ok {
			obj := t.newObject()
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), obj)
			if err != nil {
				return nil, err
			}
			objs[i] = obj
		} else {
			objs[i] = uobj
		}
	}
	return objs, nil
}

var once sync.Once

func StartCache(dynClient dynamicclient.Interface) *Listers {
	var c *Listers = nil
	once.Do(func() {
		f := NewInformerFactory(dynClient)
		c = &Listers{
			listerMap: make(map[string]cache.GenericLister),
			objectMap: make(map[string]func() runtime.Object),
		}
		stopCh := make(chan struct{})
		//scheme := runtime.NewScheme()
		//clientgoscheme.AddToScheme(scheme)
		for _, g := range gvs {
			klog.Infof("add group version resource %+v to cache", g.gv.WithResource(g.resource))
			/*
				obj, err := scheme.New(g.gvk)
				if err != nil {
					klog.Errorf("cannot get gvk: %+v", err)
					return
				}
			*/
			lister := NewLister(f, g.gv.WithResource(g.resource), stopCh)
			c.listerMap[g.resource] = lister
			c.objectMap[g.resource] = g.obj
		}
		res := f.WaitForCacheSync(stopCh)
		klog.Infof("cache synced result: %+v", res)
	})
	if c == nil {
		panic("cannot create more than one cache")
	}
	return c
}
