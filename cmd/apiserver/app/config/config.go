package config

import (
	apiserver "k8s.io/apiserver/pkg/server"
	dynamicclient "k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	"github.com/chinamobile/nlpt/apiserver/cache"
	"github.com/chinamobile/nlpt/apiserver/concurrency"
	"github.com/chinamobile/nlpt/apiserver/database"
	"github.com/chinamobile/nlpt/pkg/audit"
	dw "github.com/chinamobile/nlpt/pkg/datawarehouse"
)

type ErrorConfig struct {
	Common           map[string]string `json:"common"`
	Api              map[string]string `json:"api"`
	Application      map[string]string `json:"app"`
	ApplicationGroup map[string]string `json:"appGroup"`
	Apply            map[string]string `json:"apply"`
	DataService      map[string]string `json:"dataService"`
	DataSource       map[string]string `json:"dataSource"`
	Restriction      map[string]string `json:"restriction"`
	Serviceunit      map[string]string `json:"serviceunit"`
	ServiceunitGroup map[string]string `json:"serviceunitGroup"`
	Topic            map[string]string `json:"topic"`
	TopicGroup       map[string]string `json:"topicGroup"`
	Trafficcontrol   map[string]string `json:"trafficcontrol"`
	ClientAuth       map[string]string `json:"clientAuth"`
	ApiGroup         map[string]string `json:"apiGroup"`
}
type Config struct {
	SecureServing   *apiserver.SecureServingInfo
	InsecureServing *apiserver.DeprecatedInsecureServingInfo
	Authentication  apiserver.AuthenticationInfo
	Authorization   apiserver.AuthorizationInfo

	Client  *clientset.Clientset
	Dynamic dynamicclient.Interface

	Kubeconfig *restclient.Config

	DatasourceConfig *DatasourceConfig

	DataserviceConnector dw.Connector

	TopicConfig *TopicConfig

	Auditor *audit.Auditor

	TenantEnabled bool

	SyncMode bool

	LocalConfig ErrorConfig

	Mutex   concurrency.Mutex
	Elector concurrency.Elector

	Listers *cache.Listers

	Database *database.DatabaseConnection
}

func (c *Config) GetKubeClient() *clientset.Clientset {
	return c.Client
}

func (c *Config) GetDynamicClient() dynamicclient.Interface {
	return c.Dynamic
}

func (c *Config) GetListers() *cache.Listers {
	return c.Listers
}
