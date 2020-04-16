package api

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	k8s "github.com/chinamobile/nlpt/apiserver/kubernetes"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/crds/api/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const METRICS_NAME = "fusion_api_count_total"

type Manager struct {
	client     dynamic.NamespaceableResourceInterface
	kubeClient *clientset.Clientset
	CountDesc  *prometheus.Desc
}

func (c *Manager) List(namespace string) (*v1.ApiList, error) {
	crd, err := c.client.Namespace(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	items := &v1.ApiList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), items); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	return items, nil
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
			namespacedApiCount[ns] = len(list.Items)
		}(n)
	}
	wg.Wait()
	return
}

func (c *Manager) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.CountDesc
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
}

func NewManager(instance string, client dynamic.Interface, kubeClient *clientset.Clientset) *Manager {
	return &Manager{
		kubeClient: kubeClient,
		client:     client.Resource(v1.GetOOFSGVR()),
		CountDesc: prometheus.NewDesc(
			METRICS_NAME,
			"Number of APIs.",
			[]string{"namespace"},
			prometheus.Labels{"instance": instance},
		),
	}
}

func InitMetrics(cfg *config.Config) prometheus.Collector {
	return NewManager("nlpt", cfg.GetDynamicClient(), cfg.GetKubeClient())
}
