package handler

import (
	"github.com/emicklei/go-restful"
)

type installer interface {
	Install(ws *restful.WebService)
}
