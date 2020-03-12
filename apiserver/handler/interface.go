package handler

import (
	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type installer interface {
	Install(ws *restful.WebService)
}

type importInstaller interface {
	InstallImport(ws *restful.WebService)
}
