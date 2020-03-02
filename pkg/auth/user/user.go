package user

import (
	"strings"

	"k8s.io/klog"
)

// IMPORTANT
// ID must match regex ([A-Za-z0-9][-A-Za-z0-9_.]*) as k8s label
// At the same time, ID in auth center follow this rule
// We use ID in auth center as id here and we get names from auth center

type Users struct {
	Owner    User   `json:"owner"`
	Managers []User `json:"managers"`
	Members  []User `json:"members"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	//TenantID string `json:"tenantId"`
}

const (
	labelPrefix  = "nlpt.cmcc.com/user."
	ownerLabel   = "nlpt.cmcc.com/owner"
	managerLabel = "nlpt.cmcc.com/managers"
)

func userLabel(id string) string { return labelPrefix + id }

func InitWithOwner(id string) Users {
	return Users{
		Owner: User{
			ID: id,
		},
	}
}

func AddUsersLabels(u Users, labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}
	// owner
	labels[ownerLabel] = u.Owner.ID
	labels[userLabel(u.Owner.ID)] = "true"

	// manager
	labels[managerLabel] = func(us []User) string {
		// us: user slice
		// ul: user name list
		// ur: user
		ul := make([]string, len(us))
		for i, ur := range us {
			ul[i] = ur.ID
		}
		return strings.Join(ul, ".")
	}(u.Managers)
	for _, manager := range u.Managers {
		labels[userLabel(manager.ID)] = "true"
	}

	// member
	for _, member := range u.Members {
		labels[userLabel(member.ID)] = "true"
	}
	return labels
}

var exampleIdNames = map[string]string{
	"admin": "admin",
	"demo":  "demo",
}

func getIdNamesMap() map[string]string {
	return exampleIdNames
}

func idnames(id string) string {
	if name, ok := getIdNamesMap()[id]; ok {
		return name
	}
	return id
}

func GetUsersFromLabels(labels map[string]string) Users {
	klog.V(5).Infof("get user from label: %+v", labels)
	u := Users{
		Managers: make([]User, 0),
		Members:  make([]User, 0),
	}
	var owner string
	if u, ok := labels[ownerLabel]; ok {
		owner = u
	}
	var managers []string
	if ul, ok := labels[managerLabel]; ok {
		managers = strings.Split(ul, ".")
	}
	for k, _ := range labels {
		if len(k) > len(labelPrefix) && k[:len(labelPrefix)] == labelPrefix {
			name := k[len(labelPrefix):]
			if name == owner {
				u.Owner.ID = name
				u.Owner.Name = idnames(name)
			} else if contains(managers, name) {
				u.Managers = append(u.Managers, User{
					ID:   name,
					Name: idnames(name),
				})
			} else {
				u.Members = append(u.Members, User{
					ID:   name,
					Name: idnames(name),
				})
			}
		}
	}
	return u
}

func GetLabelSelector(id string) string {
	return userLabel(id) + "=true"
}

func contains(ss []string, t string) bool {
	for _, s := range ss {
		if t == s {
			return true
		}
	}
	return false
}
