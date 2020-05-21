package model

import (
	"github.com/chinamobile/nlpt/pkg/auth/user"
)

type UserRelation struct {
	Id           int
	ResourceType string
	ResourceId   string
	UserId       string
	Role         string
}

const (
	owner   = "owner"
	manager = "manager"
	member  = "member"
)

func FromUser(rt string, rid string, labels map[string]string) []UserRelation {
	u := user.GetUsersFromLabels(labels)
	rls := make([]UserRelation, 1+len(u.Managers)+len(u.Members))
	for i := range rls {
		rls[i].ResourceType = rt
		rls[i].ResourceId = rid
	}
	rls[0].UserId = u.Owner.ID
	rls[0].Role = owner
	for i := range u.Managers {
		rls[i+1].UserId = u.Managers[i].ID
		rls[i+1].Role = manager
	}
	ms := len(u.Managers)
	for i := range u.Members {
		rls[i+1+ms].UserId = u.Members[i].ID
		rls[i+1+ms].Role = member
	}
	return rls
}

func Equal(u1, u2 []UserRelation) bool {
	if len(u1) != len(u2) {
		return false
	}
	return true
}
