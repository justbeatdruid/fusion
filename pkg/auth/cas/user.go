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

func (u User) Less(than btree.Item) bool {
	return u.UserID < than.(User).UserID
}

type Operator interface {
	GetUserByID(string) (User, error)
	ListUsers() ([]User, error)
}

var op Operator

var once sync.Once

var cache *btree.BTree

var a int

func SetConnectionInfo(typpe string, host string, port int) {
	once.Do(func() {
		casHost = host
		casPort = port

		switch typpe {
		case "cas":
			op = NewCasOperator()
		case "tenant":
			op = NewTenantOperator()
		default:
			klog.Fatalf("unknown user backend type: %s. expect %s and %s", typpe, "cas", "tenant")
		}

		cache = btree.New(32)

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
	return "", fmt.Errorf("user id not found")
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
