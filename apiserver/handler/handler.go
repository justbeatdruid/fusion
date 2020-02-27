package handler

import (
	"github.com/chinamobile/nlpt/apiserver/resources/restriction"
	"github.com/chinamobile/nlpt/apiserver/resources/trafficcontrol"
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/chinamobile/nlpt/apiserver/handler/filter"
	"github.com/chinamobile/nlpt/apiserver/resources/api"
	"github.com/chinamobile/nlpt/apiserver/resources/application"
	"github.com/chinamobile/nlpt/apiserver/resources/applicationgroup"
	"github.com/chinamobile/nlpt/apiserver/resources/apply"
	"github.com/chinamobile/nlpt/apiserver/resources/dataservice"
	"github.com/chinamobile/nlpt/apiserver/resources/datasource"
	"github.com/chinamobile/nlpt/apiserver/resources/serviceunit"
	"github.com/chinamobile/nlpt/apiserver/resources/serviceunitgroup"
	"github.com/chinamobile/nlpt/apiserver/resources/topic"
	"github.com/chinamobile/nlpt/apiserver/resources/topicgroup"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"


	"k8s.io/apiserver/pkg/server/healthz"
)

type Handler struct {
	config *config.Config
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{cfg}
}

func (h *Handler) CreateHTTPAPIHandler(checks ...healthz.HealthChecker) (http.Handler, error) {
	wsContainer := restful.NewContainer()
	wsContainer.EnableContentEncoding(true)

	filters := []restful.FilterFunction{
		filter.NewOptionsFilter(wsContainer).Filter,
		//filter.TokenFilter{}.Filter,
	}
	for _, f := range filters {
		wsContainer.Filter(f)
	}

	healthz.InstallHandler(wsContainer, checks...)

	apiV1Ws := new(restful.WebService)

	apiV1Ws.Path("/api/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
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
	}

	for _, routerHandler := range handlers {
		routerHandler.Install(apiV1Ws)
	}

	applicationgroupHandler := applicationgroup.NewRouter(h.config)
	applicationgroupHandler.Install(apiV1Ws)

	return wsContainer, nil
}
