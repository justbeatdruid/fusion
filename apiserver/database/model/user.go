package model

import (
	"sort"

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
	var url1 UserRelationList = u1
	var url2 UserRelationList = u2
	sort.Sort(url1)
	sort.Sort(url2)
	for i := 0; i < len(u1); i++ {
		if url1[i].UserId != url2[i].UserId {
			return false
		}
		if url1[i].Role != url2[i].Role {
			return false
		}
	}
	return true
}

type UserRelationList []UserRelation

func (url UserRelationList) Len() int {
	return len(url)
}

func (url UserRelationList) Less(i, j int) bool {
	return url[i].UserId+"/"+url[i].Role < url[j].UserId+"/"+url[j].Role
}

func (url UserRelationList) Swap(i, j int) {
	url[i], url[j] = url[j], url[i]
}
