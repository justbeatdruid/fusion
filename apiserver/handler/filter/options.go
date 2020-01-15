package filter

import (
	//"strings"

	"github.com/emicklei/go-restful"
	"log"
)

type OptionsFilter struct {
	c *restful.Container
}

func (o *OptionsFilter) getContainer() *restful.Container {
	return o.c
}

func NewOptionsFilter(c *restful.Container) *OptionsFilter {
	return &OptionsFilter{
		c: c,
	}
}

// ported from go-restful/options_filter.go
// add allowed all headers, credentials
func (o *OptionsFilter) Filter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	if "OPTIONS" != req.Request.Method {
		chain.ProcessFilter(req, resp)
		return
	}

	archs := req.Request.Header.Get(restful.HEADER_AccessControlRequestHeaders)
	//methods := strings.Join(o.getContainer().ComputeAllowedMethods(req), ",")
	origin := req.Request.Header.Get(restful.HEADER_Origin)
	//if len(origin) == 0 {
	//	origin = "*"
	//}
	log.Printf("request >> origin：%s\n",origin)
	resp.AddHeader(restful.HEADER_Allow, "*")
	resp.AddHeader(restful.HEADER_AccessControlAllowOrigin, "*")
	resp.AddHeader(restful.HEADER_AccessControlAllowHeaders, archs)
	resp.AddHeader(restful.HEADER_AccessControlAllowMethods, "*")
	resp.AddHeader(restful.HEADER_AccessControlAllowCredentials, "true")
}
