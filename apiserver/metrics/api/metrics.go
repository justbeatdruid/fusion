package api

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/chinamobile/nlpt/apiserver/cache"
	k8s "github.com/chinamobile/nlpt/apiserver/kubernetes"
	"github.com/chinamobile/nlpt/crds/api/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const METRICS_NAME = "fusion_api_count_total"
const API_METRICS_NAME = "fusion_api_call_count"

type Manager struct {
	lister     *cache.ApiLister
	kubeClient *clientset.Clientset
	CountDesc  *prometheus.Desc
	apiCountDesc  *prometheus.Desc
}

func (c *Manager) List(namespace string) ([]*v1.Api, error) {
	apis, err := c.lister.List(namespace, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	return apis, nil
}

func (c *Manager) GetCount() (namespacedApiCount map[string]int) {
	namespacedApiCount = make(map[string]int)
	nss, err := k8s.GetAllNamespaces(c.kubeClient)
	if err != nil {
		klog.Errorf("cannot get namespaces: %+v", err)
		return
	}
	wg := sync.WaitGroup{}
	l := sync.Mutex{}
	wg.Add(len(nss))
	for _, n := range nss {
		go func(ns string) {
			defer wg.Done()
			list, err := c.List(ns)
			if err != nil {
				klog.Errorf("list api error: %+v", err)
				return
			}
			l.Lock()
			defer l.Unlock()
			namespacedApiCount[ns] = len(list)
		}(n)
	}
	wg.Wait()
	return
}

func (c *Manager) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.CountDesc
	ch <- c.apiCountDesc
}

func (c *Manager) Collect(ch chan<- prometheus.Metric) {
	namespacedApiCount := c.GetCount()
	for namespace, count := range namespacedApiCount {
		ch <- prometheus.MustNewConstMetric(
			c.CountDesc,
			prometheus.CounterValue,
			float64(count),
			namespace,
		)
	}
	for namespace, _ := range namespacedApiCount {
		list, err := c.List(namespace)
		if err != nil {
			klog.Errorf("list api error: %+v", err)
			return
		}
		for _, api := range list {
			ch <- prometheus.MustNewConstMetric(
				c.apiCountDesc,
				prometheus.CounterValue,
				float64(api.Status.CalledCount),
				namespace,
				api.ObjectMeta.Name,
				api.Spec.Name,
			)
		}
	}

}

func NewManager(instance string, listers *cache.Listers, kubeClient *clientset.Clientset) *Manager {
	return &Manager{
		kubeClient: kubeClient,
		lister:     listers.ApiLister(),
		CountDesc: prometheus.NewDesc(
			METRICS_NAME,
			"Number of APIs.",
			[]string{"namespace"},
			prometheus.Labels{"instance": instance},
		),
		apiCountDesc: prometheus.NewDesc(
			API_METRICS_NAME,
			"Number call count of APIs.",
			[]string{"namespace","apiId", "apiName"},
			prometheus.Labels{"instance": instance},
		),
	}
}

func InitMetrics(listers *cache.Listers, kubeClient *clientset.Clientset) prometheus.Collector {
	return NewManager("nlpt", listers, kubeClient)
}
