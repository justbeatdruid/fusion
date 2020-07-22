package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chinamobile/nlpt/apiserver/database/model"
)

type Product struct {
	// basic information
	Id            string   `json:"id"`
	Tenant        string   `json:"tenant"`
	User          string   `json:"user"`
	Category      string   `json:"category"`
	Title         string   `json:"title"`
	ApiId         string   `json:"apiId"`
	Abstract      string   `json:"abstract"` // simple abstract(char 100)
	AbstractImage string   `json:"abstractImage"`
	Appendixes    []string `json:"appendixes"` // appendix urls

	// detail
	Introduction       string   `json:"introduction"` // char 1024
	IntroductionImages []string `json:"introductionImageUrls"`

	Scenarios []Scenario `json:"scenatios"`

	CreatedAt           time.Time `json:"createdAt"`
	CreatedAtTimestamp  int64     `json:"createdAtTimestamp"`
	ReleasedAt          time.Time `json:"releasedAt"`
	ReleasedAtTimestamp int64     `json:"releasedAtTimestamp"`

	Status string `json:"status"`
}

type Scenario struct {
	Id       int    `json:"id"`
	Abstract string `json:"abstract"`
	ImageUrl string `json:"imageUrl"`
}

func FromModel(m model.Product, ss []model.Scenario) (Product, error) {
	result := Product{
		Id:            m.Id,
		Tenant:        m.Tenant,
		User:          m.User,
		Category:      m.Category,
		Title:         m.Title,
		ApiId:         m.ApiId,
		Abstract:      m.Abstract,
		AbstractImage: m.AbstractImage,

		Introduction: m.Introduction,

		CreatedAt:           m.CreatedAt,
		CreatedAtTimestamp:  m.CreatedAt.Unix(),
		ReleasedAt:          m.ReleasedAt,
		ReleasedAtTimestamp: m.ReleasedAt.Unix(),

		Status: m.Status,
	}
	if ss != nil {
		scenarios := make([]Scenario, len(ss))
		for i := range ss {
			scenarios[i] = FromModelScenario(ss[i])
		}
		result.Scenarios = scenarios
	}

	appendixes := make([]string, 0)
	if err := json.Unmarshal([]byte(m.AppendixesJson), &appendixes); err != nil {
		return Product{}, fmt.Errorf("cannot unmarshal appendixes json: %+v", err)
	}
	result.Appendixes = appendixes

	introductionImages := make([]string, 0)
	if err := json.Unmarshal([]byte(m.IntroductionImagesJson), &introductionImages); err != nil {
		return Product{}, fmt.Errorf("cannot unmarshal introductionImages json: %+v", err)
	}
	result.IntroductionImages = introductionImages

	return result, nil
}

func ToModel(a Product) (model.Product, []model.Scenario, error) {
	scenarios := make([]model.Scenario, len(a.Scenarios))
	for i := range a.Scenarios {
		scenarios[i] = ToModelScenario(a.Scenarios[i], a.Id)
	}
	result := model.Product{
		Id:            a.Id,
		Tenant:        a.Tenant,
		User:          a.User,
		Category:      a.Category,
		Title:         a.Title,
		ApiId:         a.ApiId,
		Abstract:      a.Abstract,
		AbstractImage: a.AbstractImage,

		Introduction: a.Introduction,

		CreatedAt:  a.CreatedAt,
		ReleasedAt: a.ReleasedAt,

		Status: a.Status,
	}

	var err error
	appendixes := make([]byte, 0)
	if appendixes, err = json.Marshal(&a.Appendixes); err != nil {
		return model.Product{}, nil, fmt.Errorf("cannot marshal appendixes: %+v", err)
	}
	result.AppendixesJson = string(appendixes)

	introductionImages := make([]byte, 0)
	if introductionImages, err = json.Marshal(&a.IntroductionImages); err != nil {
		return model.Product{}, nil, fmt.Errorf("cannot marshal introductionImages: %+v", err)
	}
	result.IntroductionImagesJson = string(introductionImages)

	return result, scenarios, nil
}

func FromModelScenario(m model.Scenario) Scenario {
	return Scenario{
		Id:       m.Id,
		Abstract: m.Abstract,
		ImageUrl: m.ImageUrl,
	}
}

func ToModelScenario(a Scenario, productId string) model.Scenario {
	return model.Scenario{
		Id:        a.Id,
		ProductId: productId,
		Abstract:  a.Abstract,
		ImageUrl:  a.ImageUrl,
	}
}
