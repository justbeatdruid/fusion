package service

import (
	"fmt"

	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/crds/application/api/v1"
	"github.com/chinamobile/nlpt/pkg/util"

	"k8s.io/klog"
)

func (s *Service) ListFromDatabase(opts ...util.OpOption) (*v1.ApplicationList, error) {
	op := util.OpList(opts...)
	md := &model.Application{
		Group:     op.Group(),
		Namespace: op.Namespace(),
		Name:      op.NameLike(),
	}
	uid := op.User()
	if s.tenantEnabled {
		uid = ""
	}
	mapps, err := s.db.QueryApplication(uid, md)
	if err != nil {
		return nil, fmt.Errorf("query application from database error: %+v", err)
	}
	apps := []v1.Application{}
	for _, mapp := range mapps {
		app, e := model.ApplicationToApi(mapp)
		if err != nil {
			klog.Errorf("get application error: %+v", e)
			continue
		}
		apps = append(apps, *app)
	}
	return &v1.ApplicationList{Items: apps}, nil
}
