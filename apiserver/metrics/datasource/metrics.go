package datasource

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/chinamobile/nlpt/apiserver/cache"
	k8s "github.com/chinamobile/nlpt/apiserver/kubernetes"
	"github.com/chinamobile/nlpt/crds/datasource/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const METRICS_NAME = "fusion_datasource_count_total"

type Manager struct {
	lister     *cache.DatasourceLister
	kubeClient *clientset.Clientset
	CountDesc  *prometheus.Desc
}

func (c *Manager) List(namespace string) ([]*v1.Datasource, error) {
	datasources, err := c.lister.List(namespace, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	return datasources, nil
}

func (c *Manager) GetCount() (namespacedDatasourceCount map[string]int) {
	namespacedDatasourceCount = make(map[string]int)
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
				klog.Errorf("list datasource error: %+v", err)
				return
			}
			l.Lock()
			defer l.Unlock()
			namespacedDatasourceCount[ns] = len(list)
		}(n)
	}
	wg.Wait()
	return
}

func (c *Manager) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.CountDesc
}

func (c *Manager) Collect(ch chan<- prometheus.Metric) {
	namespacedDatasourceCount := c.GetCount()
	for namespace, count := range namespacedDatasourceCount {
		ch <- prometheus.MustNewConstMetric(
			c.CountDesc,
			prometheus.CounterValue,
			float64(count),
			namespace,
		)
	}
}

func NewManager(instance string, listers *cache.Listers, kubeClient *clientset.Clientset) *Manager {
	return &Manager{
		kubeClient: kubeClient,
		lister:     listers.DatasourceLister(),
		CountDesc: prometheus.NewDesc(
			METRICS_NAME,
			"Number of datasources.",
			[]string{"namespace"},
			prometheus.Labels{"instance": instance},
		),
	}
}

func InitMetrics(listers *cache.Listers, kubeClient *clientset.Clientset) prometheus.Collector {
	return NewManager("nlpt", listers, kubeClient)
}
