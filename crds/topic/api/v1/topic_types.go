/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"strings"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TopicSpec defines the desired state of Topic
type TopicSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Topic. Edit Topic_types.go to remove/update
	Name          string        `json:"name"`
	TopicGroup    string        `json:"topicGroup"`   //topic分组ID
	PartitionNum  int           `json:"partitionNum"` //topic的分区数量，partitioned为true时，需要指定。默认为1
	OldPartitionNum int         `json:"oldPartitionNum"` //保存修改前的分区数，用于更新失败回滚
	Partitioned   bool          `json:"partitioned"`  //是否多分区，默认为false。true：代表多分区Topic
	Persistent    bool          `json:"persistent"`   //是否持久化，默认为true，非必填
	Url           string        `json:"url"`          //Topic url
	Permissions   []Permission  `json:"permissions"`
	Stats         Stats         `json:"stats"`        //Topic的统计数据
	Applications  []Application `json:"applications"` //已绑定的应用ID列表
	Description   string        `json:"description"`  //描述
	DisplayStatus ShowStatus    `json:"disStatus"`
}

type Application struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
}
type Actions []string

const (
	Consume = "consume"
	Produce = "produce"
)

type Permission struct {
	AuthUserID   string           `json:"authUserId"`   //对应clientauth的ID
	AuthUserName string           `json:"authUserName"` //对应clientauth的NAME
	Actions      Actions          `json:"actions"`      //授权的操作：发布、订阅或者发布+订阅
	Status       PermissionStatus `json:"status"`       //用户的授权状态，已授权、待删除、待授权
}

const (
	Granted = "granted" //已授权
	Grant   = "grant"
)

// TopicStatus defines the observed state of Topic
type TopicStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status              Status `json:"status"`
	Message             string `json:"message"`
	AuthorizationStatus Status `json:"authorizationStatus"`
	BindStatus          Status `json:"bindStatus"`
}

type PermissionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status  Status `json:"status"`
	Message string `json:"message"`
}

type Status string

const (
	//Init     Status = "init"
	Creating     Status = "creating"
	Created      Status = "created"
	CreateFailed Status = "createFailed"
	Deleting     Status = "deleting"
	DeleteFailed Status = "deleteFailed"
	//Error    Status = "error"
	Updating                  Status = "updating"
	Updated                   Status = "updated"
	UpdateFailed              Status = "updateFailed"
	AuthorizeFailed           Status = "authorizeFailed"
	Authorizing               Status = "authorizing"
	Authorized                Status = "authorized"
	DeletingAuthorization            = "deletingAuthorization"
	DeleteAuthorizationFailed        = "deleteAuthorizationFailed"
	DeletedAuthorization             = "deletedAuthorization"
	BindingOrUnBinding               = "bindingOrUnbinding"
	Binding                   Status = "binding"
	Bound                     Status = "bound"
	BindFailed                Status = "bindFailed"
	Unbinding                 Status = "unBinding"
	UnbindFailed              Status = "unBindFailed"
	UnbindSuccess             Status = "unBindSuccess"
)

type ShowStatus string

const (
	CreatingOfShow                  ShowStatus = "创建中"
	CreatedOfShow                   ShowStatus = "创建成功"
	CreateFailedOfShow              ShowStatus = "创建失败"
	UpdatingOfShow                  ShowStatus = "更新中"
	UpdatedOfShow                   ShowStatus = "更新成功"
	UpdateFailedOfShow              ShowStatus = "更新失败"
	DeletingOfShow                  ShowStatus = "删除中"
	DeleteFailedOfShow              ShowStatus = "删除失败"
	AuthorizingOfShow               ShowStatus = "授权中"
	AuthorizeFailedOfShow           ShowStatus = "授权失败"
	AuthorizedOfShow                ShowStatus = "授权成功"
	DeletingAuthorizationOfShow     ShowStatus = "删除授权中"
	DeleteAuthorizationFailedOfShow ShowStatus = "删除授权失败"
	DeletedAuthorizationOfShow      ShowStatus = "删除授权成功"
)

// +kubebuilder:object:root=true

// Topic is the Schema for the topics API
type Topic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopicSpec   `json:"spec,omitempty"`
	Status TopicStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopicList contains a list of Topic
type TopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Topic `json:"items"`
}

type Stats struct {
	MsgRateIn           string                      `json:"msgRateIn"`
	MsgRateOut          string                      `json:"msgRateOut"`
	MsgThroughputIn     string                      `json:"msgThroughputIn"`
	MsgThroughputOut    string                      `json:"msgThroughputOut"`
	MsgInCounter        int64                       `json:"MsgInCounter"`
	AverageMsgSize      string                      `json:"averageMsgSize"`
	BytesInCounter      int64                       `json:"bytesInCounter"`
	StorageSize         int64                       `json:"storageSize"`
	BacklogSize         int64                       `json:"backlogSize"`
	DeduplicationStatus string                      `json:"deduplicationStatus"`
	Subscriptions       map[string]SubscriptionStat `json:"subscriptions"`
	Publishers          []Publisher                 `json:"publishers"`
}
type Publisher struct {
	MsgRateIn       string `json:"msgRateIn"`
	MsgThroughputIn string `json:"msgThroughputIn"`
	AverageMsgSize  string `json:"averageMsgSize"`
	ProducerId      int64  `json:"producerId"`
	ProducerName    string `json:"producerName"`
	Address         string `json:"address"`
	ConnectedSince  string `json:"connectedSince"`
}
type SubscriptionStat struct {
	MsgRateOut                       string         `json:"msgRateOut"`
	MsgThroughputOut                 string         `json:"msgThroughputOut"`
	MsgRateRedeliver                 string         `json:"msgRateRedeliver"`
	MsgBacklog                       int64          `json:"msgBacklog"`
	BlockedSubscriptionOnUnackedMsgs bool           `json:"blockedSubscriptionOnUnackedMsgs"`
	MsgDelayed                       int64          `json:"msgDelayed"`
	UnackedMessages                  int64          `json:"unackedMessages"`
	Type                             string         `json:"type"`
	MsgRateExpired                   string         `json:"msgRateExpired"`
	LastExpireTimestamp              int64          `json:"lastExpireTimestamp"`
	LastConsumedFlowTimestamp        int64          `json:"lastConsumedFlowTimestamp"`
	LastConsumedTimestamp            int64          `json:"lastConsumedTimestamp"`
	LastAckedTimestamp               int64          `json:"lastAckedTimestamp"`
	Consumers                        []ConsumerStat `json:"consumers"`
	IsReplicated                     bool           `json:"isReplicated"`
}

type ConsumerStat struct {
	MsgRateOut string `json:"msgRateOut"`
}

func init() {
	SchemeBuilder.Register(&Topic{}, &TopicList{})
	initShowStatusMap()
}

var ShowStatusMap = make(map[Status]ShowStatus)

func (in *Topic) GetUrl() (url string) {

	var build strings.Builder
	if in.Spec.Persistent {
		build.WriteString("persistent://")
	} else {
		build.WriteString("non-persistent://")
	}

	build.WriteString(in.Namespace)
	build.WriteString("/")
	build.WriteString(in.Spec.TopicGroup)
	build.WriteString("/")
	build.WriteString(in.Spec.Name)

	return build.String()
}

func initShowStatusMap() {
	ShowStatusMap[Creating] = CreatingOfShow
	ShowStatusMap[CreateFailed] = CreateFailedOfShow
	ShowStatusMap[Created] = CreatedOfShow
	ShowStatusMap[Updating] = UpdatingOfShow
	ShowStatusMap[Updated] = UpdatedOfShow
	ShowStatusMap[UpdateFailed] = UpdateFailedOfShow
	ShowStatusMap[Deleting] = DeletingOfShow
	ShowStatusMap[DeleteFailed] = DeleteFailedOfShow
    ShowStatusMap[Authorizing] = AuthorizingOfShow
    ShowStatusMap[Authorized] = AuthorizedOfShow
    ShowStatusMap[AuthorizeFailed] = AuthorizeFailedOfShow
}
