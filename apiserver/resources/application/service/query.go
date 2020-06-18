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

func (s *Service) ListByRelationFromDatabase(resourceType, resourceId string, opts ...util.OpOption) ([]*Application, error) {
	if !s.tenantEnabled {
		return nil, fmt.Errorf("unspported for apiserver with tenant disabled")
	}
	m := &model.Application{}
	sqlTpl := `SELECT * FROM %s WHERE namespace = "%s" AND id IN (SELECT source_id FROM relation WHERE source_type = "%s" AND target_type = "%s" AND target_id = "%s")`
	sql := fmt.Sprintf(sqlTpl, m.TableName(), util.OpList(opts...).Namespace(), m.ResourceType(), resourceType, resourceId)
	mresult := make([]model.Application, 0)
	_, err := s.db.Raw(sql).QueryRows(&mresult)
	if err != nil {
		return nil, fmt.Errorf("query from database error: %+v", err)
	}
	apps := make([]*Application, len(mresult))
	for i := range mresult {
		v1app, err := model.ApplicationToApi(mresult[i])
		if err != nil {
			return nil, fmt.Errorf("get application from model error: %+v", err)
		}
		apps[i] = ToModel(v1app, opts...)
	}
	return apps, nil
}
