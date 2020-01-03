package handler

import (
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/chinamobile/nlpt/apiserver/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/apiserver/handler/filter"
	"github.com/chinamobile/nlpt/apiserver/resources/application"

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

	applicationHandler := application.NewRouter(h.config)
	applicationHandler.Install(apiV1Ws)

	/*
		kafkaHandler, err := kafka.NewHandler()
		if err != nil {
			return fmt.Errorf("cannot create kafka handler: %+v", err)
		}
		kafkaHandler.Install(apiV1Ws)
	*/

	return wsContainer, nil
}
