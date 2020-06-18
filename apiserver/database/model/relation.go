package model

import ()

type Relation struct {
	Id         int
	SourceType string
	SourceId   string
	TargetType string
	TargetId   string
}
