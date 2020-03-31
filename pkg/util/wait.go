package util

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

type Service interface {
	GetClient() dynamic.NamespaceableResourceInterface
}

func Wait(s Service, namespace string, expect func(watch.Event) bool, unexpect func(watch.Event) bool) bool {
	if expect == nil {
		return true
	}
	watcher, err := s.GetClient().Namespace(namespace).Watch(metav1.ListOptions{})
	if err != nil {
		return true
	}
	defer watcher.Stop()
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			return true
		case event := <-watcher.ResultChan():
			if _, ok := event.Object.(*unstructured.Unstructured); ok {
				if expect(event) {
					return true
				} else if unexpect != nil && unexpect(event) {
					return false
				}
			}
		}
	}
}

type Object struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

func WaitDelete(s Service, metaInfo metav1.ObjectMeta) {
	namespace := metaInfo.Namespace
	name := metaInfo.Name
	watcher, err := s.GetClient().Namespace(namespace).Watch(metav1.ListOptions{})
	if err != nil {
		return
	}
	defer watcher.Stop()
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			return
		case event := <-watcher.ResultChan():
			klog.V(5).Infof("watch an event: %+v", event)
			if event.Type != watch.Deleted {
				continue
			}
			if un, ok := event.Object.(*unstructured.Unstructured); ok {
				obj := &Object{}
				if err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), obj); err == nil {
					if obj.ObjectMeta.Name == name {
						klog.V(4).Infof("watch a delete event that matches expect name %s, ready to return request", name)
						return
					}
				} else {
					klog.Errorf("cannot cast unstructured to object: +%v", err)
				}
			} else {
				klog.Errorf("cannot case %+v to unstructured", event.Object)
			}
		}
	}
}
