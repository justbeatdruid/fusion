package model

import (
	"time"
)

type ApiGroup struct {

	// basic information
	Id          string `orm:"pk;unique"`
	Name        string `orm:"size(32)"`
	Namespace   string `orm:"size(128)"`
	User        string `orm:"size(64)"`
	Description string `orm:"size(1024)"`

	CreatedAt  time.Time `orm:"auto_now_add;type(datetime)"`
	ReleasedAt time.Time `orm:"auto_now;type(datetime)"`

	Status string `orm:"size(16)"`
}
type ApiRelation struct {
	Id         int `orm:"pk;unique;auto"`
	ApiGroupId string
	ApiId      string
}
