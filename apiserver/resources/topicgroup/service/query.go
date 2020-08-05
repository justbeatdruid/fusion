package service

import (
	"fmt"
	v1 "github.com/chinamobile/nlpt/crds/topicgroup/api/v1"

	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/pkg/util"

	"k8s.io/klog"
)


func (s *Service) ListFromDatabase(opts ...util.OpOption) (*v1.TopicgroupList, error) {
	op := util.OpList(opts...)
	md := &model.TopicGroup{
		Namespace: op.Namespace(),
		Name:      op.NameLike(),
		Available: op.Available(),
	}
	uid := op.User()

	mapps, err := s.db.QueryTopicgroup(uid, md)
	if err != nil {
		return nil, fmt.Errorf("query topicgroup from database error: %+v", err)
	}
	tgs := []v1.Topicgroup{}
	for _, mapp := range mapps {
		tg, e := model.TopicgroupToApi(mapp)
		if err != nil {
			klog.Errorf("get serviceunit error: %+v", e)
			continue
		}
		tgs = append(tgs, *tg)
	}
	return &v1.TopicgroupList{Items: tgs}, nil
}
