package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/pkg/errors"
	"github.com/chinamobile/nlpt/pkg/names"
	"regexp"
	"time"

	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/pkg/auth/cas"
)

const (
	NameReg = "^[a-zA-Z\u4e00-\u9fa5][a-zA-Z0-9_\u4e00-\u9fa5]{2,64}$"
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

func (s *Service) Validate(a *ApiGroup) error {
	for k, v := range map[string]string{
		"name":        a.Name,
		"description": a.Description,
	} {
		if k == "name" {
			if len(v) == 0 {
				return fmt.Errorf("%s is null", k)
			} else if ok, _ := regexp.MatchString(NameReg, v); !ok {
				return fmt.Errorf("name is illegal: %v", v)
			}
		}
		if k == "description" {
			if len(v) > 1024 {
				return fmt.Errorf("%s cannot exceed 1024 characters", k)
			}
		}
	}
	apiList, errs := s.ListApiGroup(*a)
	if errs != nil {
		return fmt.Errorf("cannot list apigroup object: %+v", errs)
	}
	for _, p := range apiList {
		if p.Name == a.Name {
			return errors.NameDuplicatedError("apigroup name duplicated: %+v", errs)
		}
	}
	if len(a.User) == 0 {
		return fmt.Errorf("owner not set")
	}
	a.Id = names.NewID()
	return nil
}
// target 是原始，  reqData是传进来的
func (s *Service) assignment(target *ApiGroup, reqData *ApiGroup) error {
	if len(reqData.Name) == 0 {
		return fmt.Errorf("name is nil")
	} else {
		if target.Name != reqData.Name {
			apiList, errs := s.ListApiGroup(*target)
			if errs != nil {
				return fmt.Errorf("cannot list apigroup object: %+v", errs)
			}
			for _, p := range apiList {
				if p.Name == reqData.Name {
					return errors.NameDuplicatedError("apigroup name duplicated: %+v", errs)
				}
			}
		}
		target.Name = reqData.Name
		if ok, _ := regexp.MatchString(NameReg, target.Name); !ok {
			return fmt.Errorf("name is illegal: %v ", target.Name)
		}
	}
	if len(reqData.Description) != 0 {
		if len(reqData.Description) > 1024 {
			return fmt.Errorf("%s Cannot exceed 1024 characters", reqData.Description)
		}
		target.Description = reqData.Description
	}
	return nil

}