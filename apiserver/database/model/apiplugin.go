package model

import (
	"time"
)

type ApiPlugin struct {

	// basic information
	Id          string `orm:"pk;unique"`
	Name        string `orm:"size(32)"`
	Type        string `orm:"size(64)"`
	Namespace   string `orm:"size(128)"`
	User        string `orm:"size(64)"`
	Description string `orm:"size(1024)"`

	CreatedAt  time.Time `orm:"auto_now_add;type(datetime)"`
	ReleasedAt time.Time `orm:"auto_now;type(datetime)"`
	ConsumerId string    `orm:"size(56)"`
	Raw        string    `orm:"type(text)"`

	Status string `orm:"size(16)"`
}
type ApiPluginRelation struct {
	Id           int `orm:"pk;unique;auto"`
	ApiPluginId  string
	ApiId        string
	KongPluginId string
}
