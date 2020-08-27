package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/pkg/util"
	"k8s.io/klog"

	_ "k8s.io/klog"
)

func (s *Service) ListAllApplicationApisFromDatabase(opts ...util.OpOption) ([]*Api, error) {
	sqlTpl := `select * from api where id in (select target_id from relation where source_type = 'application' and source_id in (select resource_id from user_relation where resource_type = 'application' and user_id = '%s'))`
	sql := fmt.Sprintf(sqlTpl, util.OpList(opts...).User())
	mresult := make([]model.Api, 0)
	_, err := s.db.Raw(sql).QueryRows(&mresult)
	if err != nil {
		return nil, fmt.Errorf("query from database error: %+v", err)
	}
	apis := make([]*Api, len(mresult))
	for i := range mresult {
		v1app, err := model.ApiToApi(mresult[i])
		if err != nil {
			return nil, fmt.Errorf("get application from model error: %+v", err)
		}
		apis[i] = ToModel(v1app)
	}
	return apis, nil
}

func (s *Service) ListByApiRelationFromDatabase(id string, opts ...util.OpOption) ([]*Api, error) {
	if !s.tenantEnabled {
		return nil, fmt.Errorf("unspported for apiserver with tenant disabled")
	}
	m := &model.Api{}
	sqlTpl := `SELECT * FROM %s WHERE namespace = "%s" AND id IN (SELECT api_id FROM api_relation WHERE api_group_id = "%s")`
	sql := fmt.Sprintf(sqlTpl, m.TableName(), util.OpList(opts...).Namespace(), id)
	klog.Infof("query api sql: %s", sql)
	mResult := make([]model.Api, 0)
	_, err := s.db.Raw(sql).QueryRows(&mResult)
	if err != nil {
		return nil, fmt.Errorf("query from database error: %+v", err)
	}
	apis := make([]*Api, len(mResult))
	for i := range mResult {
		v1api, err := model.ApiToApi(mResult[i])
		if err != nil {
			return nil, fmt.Errorf("get application from model error: %+v", err)
		}
		apis[i] = ToModel(v1api)
	}
	return apis, nil
}

func (s *Service) ListByApiPluginRelationFromDatabase(id string, opts ...util.OpOption) ([]*Api, error) {
	if !s.tenantEnabled {
		return nil, fmt.Errorf("unspported for apiserver with tenant disabled")
	}
	m := &model.Api{}
	sqlTpl := `SELECT * FROM %s WHERE namespace = "%s" AND id IN (SELECT target_id FROM api_plugin_relation WHERE api_plugin_id = "%s" AND target_type = "api" )`
	sql := fmt.Sprintf(sqlTpl, m.TableName(), util.OpList(opts...).Namespace(), id)
	klog.Infof("query api sql: %s", sql)
	mResult := make([]model.Api, 0)
	_, err := s.db.Raw(sql).QueryRows(&mResult)
	if err != nil {
		return nil, fmt.Errorf("query from database error: %+v", err)
	}
	apis := make([]*Api, len(mResult))
	for i := range mResult {
		v1api, err := model.ApiToApi(mResult[i])
		if err != nil {
			return nil, fmt.Errorf("get application from model error: %+v", err)
		}
		apis[i] = ToModel(v1api)
	}
	return apis, nil
}
