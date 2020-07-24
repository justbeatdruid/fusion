package model

import (
	"time"
)

type Product struct {

	// basic information
	Id             string `orm:"pk;unique"`
	Tenant         string `orm:"size(64)"`
	User           string `orm:"size(64)"`
	Category       string `orm:"size(64)"`
	Title          string `orm:"size(64)"`
	ApiId          string `orm:"size(64)"`
	Abstract       string `orm:"size(1024)"`
	AbstractImage  string `orm:"size(64)"`
	AppendixesJson string `orm:"size(1024)"`

	// detail
	Introduction           string `orm:"type(text)"`
	IntroductionImagesJson string `orm:"type(text)"`

	CreatedAt  time.Time `orm:"auto_now_add;type(datetime)"`
	ReleasedAt time.Time `orm:"auto_now;type(datetime)"`

	Status string `orm:"size(16)"`
}

type Scenario struct {
	Id        int `orm:"pk;unique"`
	ProductId string
	Abstract  string
	ImageUrl  string
}
