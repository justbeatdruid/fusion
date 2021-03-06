package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/pkg/util"
	"k8s.io/klog"
	"strings"

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

func (s *Service) ListByApiPluginRelationFromDatabase(id string, opts ...util.OpOption) ([]*ApiRes, error) {
	if !s.tenantEnabled {
		return nil, fmt.Errorf("unspported for apiserver with tenant disabled")
	}

	//sqlTpl := `SELECT * FROM %s WHERE namespace = "%s" AND id IN (SELECT target_id FROM api_plugin_relation WHERE api_plugin_id = "%s" AND target_type = "api" )`
	sqlTpl := `SELECT api.id, api.name, api_plugin_relation.status, api_plugin_relation.detail, api_plugin_relation.enable FROM api_plugin_relation LEFT JOIN api on api.id = api_plugin_relation.target_id where api_plugin_relation.api_plugin_id="%s" and api.namespace="%s"`
	sql := fmt.Sprintf(sqlTpl, id, util.OpList(opts...).Namespace())
	klog.Infof("query api sql: %s", sql)
	mResult := make([]*ApiRes, 0)
	_, err := s.db.Raw(sql).QueryRows(&mResult)
	if err != nil {
		return nil, fmt.Errorf("query from database error: %+v", err)
	}
	return mResult, nil
}

func (s *Service) ListForCapabilityFromDatabase(ids []string, opts ...util.OpOption) ([]*Api, error) {
	if !s.tenantEnabled {
		return nil, fmt.Errorf("unspported for apiserver with tenant disabled")
	}
	idStr := strings.Join(ids, "','")
	m := &model.Api{}
	sqlTpl := `SELECT * FROM %s WHERE id IN ('%s')`
	sql := fmt.Sprintf(sqlTpl, m.TableName(), idStr)
	klog.Infof("query api for capability sql: %s", sql)
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
