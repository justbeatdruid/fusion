package handler

import (
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/cache"
	"github.com/chinamobile/nlpt/apiserver/handler/callback"
	"github.com/chinamobile/nlpt/apiserver/handler/filter"
	"github.com/chinamobile/nlpt/apiserver/resources/api"
	"github.com/chinamobile/nlpt/apiserver/resources/apigroup"
	"github.com/chinamobile/nlpt/apiserver/resources/apiplugin"
	"github.com/chinamobile/nlpt/apiserver/resources/application"
	"github.com/chinamobile/nlpt/apiserver/resources/applicationgroup"
	"github.com/chinamobile/nlpt/apiserver/resources/apply"
	"github.com/chinamobile/nlpt/apiserver/resources/clientauth"
	"github.com/chinamobile/nlpt/apiserver/resources/dataservice"
	"github.com/chinamobile/nlpt/apiserver/resources/datasource"
	"github.com/chinamobile/nlpt/apiserver/resources/product"
	"github.com/chinamobile/nlpt/apiserver/resources/restriction"
	"github.com/chinamobile/nlpt/apiserver/resources/serviceunit"
	"github.com/chinamobile/nlpt/apiserver/resources/serviceunitgroup"
	"github.com/chinamobile/nlpt/apiserver/resources/topic"
	"github.com/chinamobile/nlpt/apiserver/resources/topicgroup"
	"github.com/chinamobile/nlpt/apiserver/resources/trafficcontrol"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/go-restful"
	swagger "github.com/chinamobile/nlpt/pkg/go-restful-swagger12"

	"k8s.io/apiserver/pkg/server/healthz"
)

type Handler struct {
	config *config.Config
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{cfg}
}

func (h *Handler) CreateHTTPAPIHandler(checks ...healthz.HealthChecker) (http.Handler, error) {
	if h.config.SyncMode {
		h.config.Listers = cache.StartCache(h.config.GetDynamicClient(), h.config.Database)
		return restful.NewContainer(), nil
	} else {
		h.config.Listers = cache.StartCache(h.config.GetDynamicClient(), nil)
	}
	//TODO cache will be used in service

	wsContainer := restful.NewContainer()
	wsContainer.EnableContentEncoding(true)

	filters := []restful.FilterFunction{
		filter.NewOptionsFilter(wsContainer).Filter,
		filter.NewTokenFilter().Filter,
	}
	for _, f := range filters {
		wsContainer.Filter(f)
	}

	wsContainer.Callback(callback.NewAuditCaller(wsContainer, h.config.Auditor, h.config.TenantEnabled))
	wsContainer.Callback(callback.NewLogger())

	healthz.InstallHandler(wsContainer, checks...)

	apiV1Ws := new(restful.WebService)

	apiV1Ws.Path("/api/v1")

	wsContainer.Add(apiV1Ws)

	handlers := []installer{
		api.NewRouter(h.config),
		application.NewRouter(h.config),
		applicationgroup.NewRouter(h.config),
		serviceunit.NewRouter(h.config),
		serviceunitgroup.NewRouter(h.config),
		datasource.NewRouter(h.config),
		apply.NewRouter(h.config),
		topic.NewRouter(h.config),
		trafficcontrol.NewRouter(h.config),
		dataservice.NewRouter(h.config),
		restriction.NewRouter(h.config),
		topicgroup.NewRouter(h.config),
		clientauth.NewRouter(h.config),
		product.NewRouter(h.config),
		apigroup.NewRouter(h.config),
		apiplugin.NewRouter(h.config),
	}

	for _, routerHandler := range handlers {
		routerHandler.Install(apiV1Ws)

	}

	applicationgroupHandler := applicationgroup.NewRouter(h.config)
	applicationgroupHandler.Install(apiV1Ws)

	metricsWs := new(restful.WebService)
	metricsWs.Path("/metrics")
	metricsWs.Route(metricsWs.GET("/").
		Consumes("*/*").
		Produces("*/*").
		To(NewMetricsHandler(h.config)))

	wsContainer.Add(metricsWs)

	config := swagger.Config{
		WebServices:     []*restful.WebService{apiV1Ws, metricsWs},
		ApiPath:         "/api/v1/apidocs.json",
		SwaggerPath:     "/api/v1/apidocs/",
		SwaggerFilePath: "/data/web/swagger-ui/dist"}
	swagger.RegisterSwaggerService(config, wsContainer)

	return wsContainer, nil
}
