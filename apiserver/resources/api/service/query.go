package service

import (
	"fmt"

	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/pkg/util"

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
