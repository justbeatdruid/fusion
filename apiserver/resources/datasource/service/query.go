package service

import (
	"fmt"

	"github.com/chinamobile/nlpt/apiserver/database/model"
	//"github.com/chinamobile/nlpt/crds/datasource/api/v1"
	"github.com/chinamobile/nlpt/pkg/util"
	//"k8s.io/klog"
)

func (s *Service) ListDatasourceFromDatabase(opts ...util.OpOption) ([]*Datasource, error) {
	if !s.tenantEnabled {
		return nil, fmt.Errorf("unspported for apiserver with tenant disabled")
	}
	op := util.OpList(opts...)
	md := &model.Datasource{
		Namespace: op.Namespace(),
		Name:      op.NameLike(),
		Type:      op.Type(),
		User:      op.User(),
		Status:    op.Status(),
	}
	mresult, err := s.db.QueryDatasource(md)
	if err != nil {
		return nil, fmt.Errorf("cannot query from database: %+v", err)
	}
	dss := make([]*Datasource, len(mresult))
	for i := range mresult {
		v1ds, err := model.DatasourceToApi(mresult[i])
		if err != nil {
			return nil, fmt.Errorf("get datasource from model error: %+v", err)
		}
		dss[i] = ToModel(v1ds)
	}
	return dss, nil
}
