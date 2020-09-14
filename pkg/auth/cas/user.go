package cas

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/btree"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

type User struct {
	UserID      int    `json:"userId"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Salt        string `json:"salt"`
	Status      int    `json:"status"`
	Email       string `json:"email"`
	Mobile      string `json:"mobile"`
	Description string `json:"description"`
	CreateTime  string `json:"createdAt"`
}

type Tenant struct {
	TenantID   int    `json:"tenantId"`
	TenantName string `json:"tenantName"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

func (t Tenant) Less(than btree.Item) bool {
	return t.TenantID < than.(Tenant).TenantID
}

func (u User) Less(than btree.Item) bool {
	return u.UserID < than.(User).UserID
}

type Operator interface {
	GetUserByID(string) (User, error)
	ListUsers() ([]User, error)

	GetTenantByID(string) (Tenant, error)
	ListTenants() ([]Tenant, error)
}

var op Operator

var once sync.Once

var cache *btree.BTree

var tenantCache *btree.BTree

var a, b int

func SetConnectionInfo(typpe string, host string, port int) {
	once.Do(func() {
		casHost = host
		casPort = port

		switch typpe {
		case "cas":
			op = NewCasOperator()
		case "tenant":
			op = NewFusionOperator()
		default:
			klog.Fatalf("unknown user backend type: %s. expect %s and %s", typpe, "cas", "tenant")
		}

		cache = btree.New(32)
		tenantCache = btree.New(32)

		go wait.Until(func() {
			count, err := syncUsers()
			if err != nil {
				klog.Errorf("sync user error: %+v", err)
			} else {
				if a%180 == 0 {
					klog.V(5).Infof("successfully sync %d users", count)
					a++
				}
			}
			count, err = syncTenants()
			if err != nil {
				klog.Errorf("sync tenant error: %+v", err)
			} else {
				if b%180 == 0 {
					klog.V(5).Infof("successfully sync %d tenants", count)
					b++
				}
			}
		}, time.Second*10, wait.NeverStop)
	})
}

// TODO for create(update) actions, we should directly get user from cas
func GetUserNameByID(id string) (string, error) {
	iid, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		return "", fmt.Errorf("expect user id in int format. found %s", id)
	}
	if u := cache.Get(User{
		UserID: int(iid),
	}); u != nil {
		return u.(User).Username, nil
	}
	if false {
		user, err := op.GetUserByID(id)
		return user.Username, err
	}
	return "", fmt.Errorf("user id %s not found", id)
}

func GetTenantNameByID(id string) (string, error) {
	iid, err := strconv.ParseInt(id, 10, 0)
	if err != nil {
		return "", fmt.Errorf("expect user id in int format. found %s", id)
	}
	if t := tenantCache.Get(Tenant{
		TenantID: int(iid),
	}); t != nil {
		return t.(Tenant).TenantName, nil
	}
	return "", fmt.Errorf("tenant id %s not found", id)
}

func syncUsers() (int, error) {
	users, err := op.ListUsers()
	if err != nil {
		return 0, fmt.Errorf("list user error: %+v", err)
	}
	for _, u := range users {
		cache.ReplaceOrInsert(u)
	}
	return len(users), nil
}

func syncTenants() (int, error) {
	tenants, err := op.ListTenants()
	if err != nil {
		return 0, fmt.Errorf("list user error: %+v", err)
	}
	for _, t := range tenants {
		tenantCache.ReplaceOrInsert(t)
	}
	return len(tenants), nil
}
