package model

type Table interface {
	TableName() string
	ResourceType() string
	ResourceId() string
}

type Condition struct {
	Field    string
	Operator Op
	Value    string
}

type Op int

const (
	Equals Op = iota
	LessThan
	NotLessThan
	MoreThan
	NotMoreThan
	Like
	In
)
