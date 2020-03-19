package user

import (
	"fmt"
	"strings"

	"github.com/chinamobile/nlpt/pkg/auth/cas"

	"k8s.io/klog"
)

// IMPORTANT
// ID must match regex ([A-Za-z0-9][-A-Za-z0-9_.]*) as k8s label
// At the same time, ID in auth center follow this rule
// We use ID in auth center as id here and we get names from auth center
const split = "."

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
		return strings.Join(ul, split)
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
	if name, err := cas.GetUserNameByID(id); err == nil {
		return name
	} else {
		klog.Errorf("get user name by id error: %+v, use id as name", err)
	}
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
		managers = Split(ul, split)
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

func IsOwner(id string, labels map[string]string) bool {
	return id == labels[ownerLabel] && len(id) > 0
}

func IsManager(id string, labels map[string]string) bool {
	managers := Split(labels[managerLabel], split)
	return contains(managers, id)
}

func IsUser(id string, labels map[string]string) bool {
	_, ok := labels[userLabel(id)]
	return ok
}

func GetOwner(labels map[string]string) string {
	return labels[ownerLabel]
}

var ReadPermitted = IsUser

func WritePermitted(id string, labels map[string]string) bool {
	return IsManager(id, labels) || IsOwner(id, labels)
}

func AddUserLabels(d *Data, labels map[string]string) (map[string]string, error) {
	if labels == nil {
		labels = make(map[string]string)
	}
	if d == nil {
		return labels, fmt.Errorf("data is null")
	}
	if len(d.ID) == 0 {
		return labels, fmt.Errorf("id in data is null")
	}
	if _, ok := labels[userLabel(d.ID)]; ok {
		return labels, fmt.Errorf("user already exists")
	}
	labels[userLabel(d.ID)] = "true"
	switch d.Role {
	case Manager:
		managers := Split(labels[managerLabel], split)
		managers = addItem(managers, d.ID)
		labels[managerLabel] = strings.Join(managers, split)
	case Member:
	default:
		return labels, fmt.Errorf("wrong user role: %s", d.Role)
	}
	return labels, nil
}

func RemoveUserLabels(id string, labels map[string]string) (map[string]string, error) {
	if labels == nil {
		labels = make(map[string]string)
	}
	if len(id) == 0 {
		return labels, fmt.Errorf("id is null")
	}
	if _, ok := labels[userLabel(id)]; !ok {
		return labels, fmt.Errorf("user not exists")
	}
	owner := labels[ownerLabel]
	if id == owner {
		return labels, fmt.Errorf("cannot remove owner")
	}
	delete(labels, userLabel(id))

	managers := Split(labels[managerLabel], split)
	managers = removeItem(managers, id)
	labels[managerLabel] = strings.Join(managers, split)
	return labels, nil
}

// original owner will be member
func ChangeOwner(id string, labels map[string]string) (map[string]string, error) {
	if labels == nil {
		labels = make(map[string]string)
	}
	if len(id) == 0 {
		return labels, fmt.Errorf("id is null")
	}
	if _, ok := labels[userLabel(id)]; !ok {
		return labels, fmt.Errorf("user not exists")
	}
	owner := labels[ownerLabel]
	if id == owner {
		return labels, fmt.Errorf("owner is already %s", id)
	}
	// 1. update owner
	labels[ownerLabel] = id
	// 2. remove member from managers
	managers := Split(labels[managerLabel], split)
	managers = removeItem(managers, id)
	labels[managerLabel] = strings.Join(managers, split)
	return labels, nil
}

func ChangeUser(d *Data, labels map[string]string) (map[string]string, error) {
	if labels == nil {
		labels = make(map[string]string)
	}
	if d == nil {
		return labels, fmt.Errorf("data is null")
	}
	if len(d.ID) == 0 {
		return labels, fmt.Errorf("id in data is null")
	}
	if _, ok := labels[userLabel(d.ID)]; !ok {
		return labels, fmt.Errorf("user not exists")
	}
	owner := labels[ownerLabel]
	if d.ID == owner {
		return labels, fmt.Errorf("cannot change owner role")
	}
	switch d.Role {
	case Manager:
		managers := Split(labels[managerLabel], split)
		managers = addItem(managers, d.ID)
		labels[managerLabel] = strings.Join(managers, split)
	case Member:
		managers := Split(labels[managerLabel], split)
		managers = removeItem(managers, d.ID)
		labels[managerLabel] = strings.Join(managers, split)
	default:
		return labels, fmt.Errorf("wrong user role: %s", d.Role)
	}
	klog.V(5).Infof("change user labels: %+v", labels)
	return labels, nil
}

func Split(s, pattern string) []string {
	if len(s) == 0 {
		return make([]string, 0)
	}
	return strings.Split(s, pattern)
}

func contains(ss []string, t string) bool {
	for _, s := range ss {
		if t == s {
			return true
		}
	}
	return false
}

func addItem(ss []string, t string) []string {
	for _, s := range ss {
		if t == s {
			return ss
		}
	}
	return append(ss, t)
}

func removeItem(ss []string, t string) []string {
	idx := -1
	for i, s := range ss {
		if t == s {
			idx = i
		}
	}
	if idx == -1 {
		return ss
	}
	return append(ss[:idx], ss[idx+1:]...)
}

type Data struct {
	ID   string `json:"id"`
	Role Role   `json:"role"`
}

type Role string

const (
	Manager = "manager"
	Member  = "member"
)

type Wrapped struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *Data  `json:"data,omitempty"`
}

type UserResponse = Wrapped
type UserRequest = Wrapped
