package service

import (
	"time"

	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/pkg/auth/cas"
)

type ApiGroup struct {
	// basic information
	Id          string `json:"id"`
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	User        string `json:"user"`
	UserName    string `json:"userName"`
	Description string `json:"description"`

	ApiRelation []ApiRelation `json:"apirelation"`

	CreatedAt           time.Time `json:"createdAt"`
	CreatedAtTimestamp  int64     `json:"createdAtTimestamp"`
	ReleasedAt          time.Time `json:"releasedAt"`
	ReleasedAtTimestamp int64     `json:"releasedAtTimestamp"`

	Status string `json:"status"`
}

type ApiRelation struct {
	Id         int    `json:"id"`
	ApiGroupId string `json:"apiGroupId"`
	ApiId      string `json:"apiId"`
}

type ApiBind struct {
	ID string `json:"id"`
}

func FromModel(m model.ApiGroup, ss []model.ApiRelation) (ApiGroup, error) {
	result := ApiGroup{
		Id:                  m.Id,
		Name:                m.Name,
		Namespace:           m.Namespace,
		User:                m.User,
		CreatedAt:           m.CreatedAt,
		CreatedAtTimestamp:  m.CreatedAt.Unix(),
		ReleasedAt:          m.ReleasedAt,
		ReleasedAtTimestamp: m.ReleasedAt.Unix(),

		Status: m.Status,
	}
	if ss != nil {
		scenarios := make([]ApiRelation, len(ss))
		for i := range ss {
			scenarios[i] = FromModelScenario(ss[i])
		}
		result.ApiRelation = scenarios
	}

	username, err := cas.GetUserNameByID(m.User)
	if err == nil {
		result.UserName = username
	} else {
		result.UserName = "用户数据错误"
	}
	return result, nil
}

func ToModel(a ApiGroup) (model.ApiGroup, []model.ApiRelation, error) {
	apis := make([]model.ApiRelation, len(a.ApiRelation))
	for i := range a.ApiRelation {
		apis[i] = ToModelScenario(a.ApiRelation[i], a.Id, "")
	}
	result := model.ApiGroup{
		Id:          a.Id,
		Name:        a.Name,
		Namespace:   a.Namespace,
		User:        a.User,
		Description: a.Description,
		CreatedAt:   a.CreatedAt,
		ReleasedAt:  a.ReleasedAt,

		Status: a.Status,
	}

	return result, apis, nil
}

func FromModelScenario(m model.ApiRelation) ApiRelation {
	return ApiRelation{
		Id:         m.Id,
		ApiGroupId: m.ApiGroupId,
		ApiId:      m.ApiId,
	}
}

func ToModelScenario(a ApiRelation, productId string, apiId string) model.ApiRelation {
	return model.ApiRelation{
		Id:         a.Id,
		ApiGroupId: productId,
		ApiId:      apiId,
	}
}
