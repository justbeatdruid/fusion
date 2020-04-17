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

	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	applicationv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	applicationgroupv1 "github.com/chinamobile/nlpt/crds/applicationgroup/api/v1"
	applyv1 "github.com/chinamobile/nlpt/crds/apply/api/v1"
	clientauthv1 "github.com/chinamobile/nlpt/crds/clientauth/api/v1"
	dataservicev1 "github.com/chinamobile/nlpt/crds/dataservice/api/v1"
	datasourcev1 "github.com/chinamobile/nlpt/crds/datasource/api/v1"
	restrictionv1 "github.com/chinamobile/nlpt/crds/restriction/api/v1"
	serviceunitv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	serviceunitgroupv1 "github.com/chinamobile/nlpt/crds/serviceunitgroup/api/v1"
	topicv1 "github.com/chinamobile/nlpt/crds/topic/api/v1"
	topicgroupv1 "github.com/chinamobile/nlpt/crds/topicgroup/api/v1"
	trafficcontrolv1 "github.com/chinamobile/nlpt/crds/trafficcontrol/api/v1"
)

type gv struct {
	resource string
	gv       schema.GroupVersion
	obj      func() runtime.Object
}

var gvs = []gv{
	{"apis", apiv1.GroupVersion, func() runtime.Object { return &apiv1.Api{} }},
	{"applications", applicationv1.GroupVersion, func() runtime.Object { return &applicationv1.Application{} }},
	{"applicationgroups", applicationgroupv1.GroupVersion, func() runtime.Object { return &applicationgroupv1.ApplicationGroup{} }},
	{"applies", applyv1.GroupVersion, func() runtime.Object { return &applyv1.Apply{} }},
	{"clientauths", clientauthv1.GroupVersion, func() runtime.Object { return &clientauthv1.Clientauth{} }},
	{"dataservices", dataservicev1.GroupVersion, func() runtime.Object { return &dataservicev1.Dataservice{} }},
	{"datasources", datasourcev1.GroupVersion, func() runtime.Object { return &datasourcev1.Datasource{} }},
	{"restrictions", restrictionv1.GroupVersion, func() runtime.Object { return &restrictionv1.Restriction{} }},
	{"serviceunits", serviceunitv1.GroupVersion, func() runtime.Object { return &serviceunitv1.Serviceunit{} }},
	{"serviceunitgroups", serviceunitgroupv1.GroupVersion, func() runtime.Object { return &serviceunitgroupv1.ServiceunitGroup{} }},
	{"topics", topicv1.GroupVersion, func() runtime.Object { return &topicv1.Topic{} }},
	{"topicgroups", topicgroupv1.GroupVersion, func() runtime.Object { return &topicgroupv1.Topicgroup{} }},
	{"trafficcontrols", trafficcontrolv1.GroupVersion, func() runtime.Object { return &trafficcontrolv1.Trafficcontrol{} }},
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
