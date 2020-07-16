package service

import (
	"fmt"

	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	"github.com/chinamobile/nlpt/pkg/util"

	"k8s.io/klog"
)

func (s *Service) ListFromDatabase(opts ...util.OpOption) (*v1.ServiceunitList, error) {
	op := util.OpList(opts...)
	md := &model.Serviceunit{
		Group:     op.Group(),
		Namespace: op.Namespace(),
		Name:      op.NameLike(),
		Type:      op.Stype(),
	}
	uid := op.User()
	if s.tenantEnabled {
		uid = ""
	}
	mapps, err := s.db.QueryServiceunit(uid, md)
	if err != nil {
		return nil, fmt.Errorf("query serviceunit from database error: %+v", err)
	}
	apps := []v1.Serviceunit{}
	for _, mapp := range mapps {
		app, e := model.ServiceunitToApi(mapp)
		if err != nil {
			klog.Errorf("get serviceunit error: %+v", e)
			continue
		}
		apps = append(apps, *app)
	}
	return &v1.ServiceunitList{Items: apps}, nil
}
