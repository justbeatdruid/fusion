package cache

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicclient "k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

func NewInformerFactory(dynClient dynamicclient.Interface) dynamicinformer.DynamicSharedInformerFactory {
	return dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynClient, 0, metav1.NamespaceAll, nil)
}

func NewLister(f dynamicinformer.DynamicSharedInformerFactory, gvr schema.GroupVersionResource, stopCh <-chan struct{}) cache.GenericLister {
	i := f.ForResource(gvr)
	go startWatching(stopCh, i.Informer(), gvr.Resource)
	return i.Lister()
}

func startWatching(stopCh <-chan struct{}, s cache.SharedIndexInformer, r string) {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			klog.V(9).Infof("received add event")
		},
		UpdateFunc: func(oldObj, obj interface{}) {
			klog.V(9).Infof("received update event")
		},
		DeleteFunc: func(obj interface{}) {
			klog.V(9).Infof("received update event")
		},
	}
	if h, ok := gvm[r]; ok {
		handlers = h
	}

	s.AddEventHandler(handlers)
	s.Run(stopCh)
}
