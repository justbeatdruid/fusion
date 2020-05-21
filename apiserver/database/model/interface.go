package model

type Table interface {
	TableName() string
	ResourceType() string
	ResourceId() string
}
