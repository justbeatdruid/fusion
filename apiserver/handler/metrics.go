package handler

import (
	"github.com/chinamobile/nlpt/apiserver/metrics/api"
	"github.com/chinamobile/nlpt/apiserver/metrics/datasource"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/go-restful"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"k8s.io/klog"
)

type promLogger struct{}

func (*promLogger) Println(v ...interface{}) {
	klog.Errorf("Prometheus: %+v", v...)
}

func NewMetricsHandler(cfg *config.Config) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		collectors := []prometheus.Collector{
			api.InitMetrics(cfg),
			datasource.InitMetrics(cfg),
		}
		reg := prometheus.NewPedanticRegistry()
		reg.MustRegister(collectors...)
		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			reg,
		}
		handler := promhttp.HandlerFor(gatherers,
			promhttp.HandlerOpts{
				ErrorLog:      &promLogger{},
				ErrorHandling: promhttp.ContinueOnError,
			})
		handler.ServeHTTP(
			resp.ResponseWriter,
			req.Request)
	}
}
