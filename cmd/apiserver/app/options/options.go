package options

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	dynamicclient "k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"

	"github.com/chinamobile/nlpt/apiserver/cache"
	"github.com/chinamobile/nlpt/apiserver/concurrency"
	"github.com/chinamobile/nlpt/apiserver/database"
	appconfig "github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/audit"
	"github.com/chinamobile/nlpt/pkg/auth/cas"
	dw "github.com/chinamobile/nlpt/pkg/datawarehouse"
)

const userAgent = "nlpt"

type ServerRunOptions struct {
	ListenAddress string
	Master        string
	Kubeconfig    string
	CrdNamespace  string
	TenantEnabled bool
	SyncMode      bool

	ConfigPath string

	Datasource  *DatasourceOptions
	Dataservice *DataserviceOptions
	Topic       *TopicOptions
	Cas         *CasOptions
	Audit       *AuditOptions
	Etcd        *EtcdOptions

	Database *DatabaseOptions
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		ListenAddress: ":8001",
		CrdNamespace:  os.Getenv("MY_POD_NAMESPACE"),
		TenantEnabled: false,
		SyncMode:      false,
		ConfigPath:    "/data/err.json",
		Datasource:    DefaultDatasourceOptions(),
		Dataservice:   DefaultDataserviceOptions(),
		Topic:         DefaultTopicOptions(),
		Cas:           DefaultCasOptions(),
		Audit:         DefaultAuditOptions(),
		Etcd:          DefaultEtcdOptions(),

		Database: DefaultDatabaseOptions(),
	}

	if len(s.CrdNamespace) == 0 {
		klog.V(4).Infof("cannot find environmnent MY_POD_NAMESPACE, use default")
		s.CrdNamespace = "default"
	}

	klog.V(5).Infof("after parse options ConfigPath %s", s.ConfigPath)
	return s
}

func (s *ServerRunOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	s.Datasource.AddFlags(fss.FlagSet("data source"))
	s.Dataservice.AddFlags(fss.FlagSet("data service"))
	s.Topic.AddFlags(fss.FlagSet("topic"))
	s.Cas.AddFlags(fss.FlagSet("cas"))
	s.Audit.AddFlags(fss.FlagSet("audit"))
	s.Etcd.AddFlags(fss.FlagSet("etcd"))
	s.Database.AddFlags(fss.FlagSet("database"))
	kfset := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(kfset)
	fss.FlagSet("klog").AddGoFlagSet(kfset)
	fs := fss.FlagSet("misc")
	fs.StringVar(&s.Master, "master", s.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig).")
	fs.StringVar(&s.Kubeconfig, "kubeconfig", s.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")
	fs.StringVar(&s.ListenAddress, "listen", s.ListenAddress, "Address of app manager server listening on")
	fs.BoolVar(&s.TenantEnabled, "tenant-enabled", s.TenantEnabled, "Enable tenant, that means for on resource, role of user is always same for all items in tenant. If tenant enabled, we do not check anymore for user's writing or reading permission")
	fs.BoolVar(&s.SyncMode, "sync-mode", s.SyncMode, "If true, run sync only and do not run rest server. If false, run rest server but not sync data from crd to database")
	fs.StringVar(&s.ConfigPath, "local-config", s.ConfigPath, "Location of local config")
	return fss
}

func (s *ServerRunOptions) ParseLocalConfig() (*appconfig.ErrorConfig, error) {
	errConfig := &appconfig.ErrorConfig{}
	if len(s.ConfigPath) == 0 {
		return nil, fmt.Errorf("config path not set")
	}
	configFile, err := os.Open(s.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %+v", err)
	}

	decoder := json.NewDecoder(configFile)
	if err = decoder.Decode(errConfig); err != nil {
		return nil, fmt.Errorf("error parsing config file: %+v", err)
	}
	klog.Infof("parse options s.LocalConfig %+v", errConfig)
	return errConfig, nil
}

func (s *ServerRunOptions) Config() (*appconfig.Config, error) {
	kubeconfig, err := clientcmd.BuildConfigFromFlags(s.Master, s.Kubeconfig)
	if err != nil {
		return nil, err
	}
	//kubeconfig.DisableCompression = true

	client, err := clientset.NewForConfig(restclient.AddUserAgent(kubeconfig, userAgent))
	if err != nil {
		return nil, err
	}

	dynClient, err := dynamicclient.NewForConfig(restclient.AddUserAgent(kubeconfig, userAgent))
	if err != nil {
		return nil, err
	}
	if s.SyncMode {
		s.Database.Enabled = true
		db, err := database.NewDatabaseConnection(s.Database.Enabled, s.Database.Type, s.Database.Host, s.Database.Port, s.Database.Username, s.Database.Password,
			s.Database.Database, s.Database.Schema)
		if err != nil {
			return nil, fmt.Errorf("cannot connect to database: %+v", err)
		}
		casType := "cas"
		if s.TenantEnabled {
			casType = "tenant"
		} else {
			cache.SetGVs([]string{"apis", "applications", "applicationgroups", "applies", "datasources", "serviceunits", "serviceunitgroups", "topics", "topicgroups"})
		}
		cas.SetConnectionInfo(casType, s.Cas.Host, s.Cas.Port)
		return &appconfig.Config{
			Client:     client,
			Dynamic:    dynClient,
			Kubeconfig: kubeconfig,
			SyncMode:   s.SyncMode,
			Database:   db,
		}, nil
	}

	errConfig, err := s.ParseLocalConfig()
	if err != nil {
		return nil, fmt.Errorf("parse error config error: %+v", err)
	}

	if err = s.Topic.ParsePulsarSecret(); err != nil {
		return nil, fmt.Errorf("parse pulsar secret error: %+v", err)
	}
	mtx, err := concurrency.NewEtcdMutex(s.Etcd.Endpoints, s.Etcd.CertFile, s.Etcd.KeyFile, s.Etcd.CAFile, s.Etcd.Timeout)
	if err != nil {
		return nil, fmt.Errorf("cannot create etcd mutex: %+v", err)
	}
	elector, err := concurrency.NewEtcdElector(s.Etcd.Endpoints, s.Etcd.CertFile, s.Etcd.KeyFile, s.Etcd.CAFile, s.Etcd.Timeout)
	if err != nil {
		return nil, fmt.Errorf("cannot create etcd elector: %+v", err)
	}
	db, err := database.NewDatabaseConnection(s.Database.Enabled, s.Database.Type, s.Database.Host, s.Database.Port, s.Database.Username, s.Database.Password,
		s.Database.Database, s.Database.Schema)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %+v", err)
	}
	c := &appconfig.Config{
		Client:     client,
		Dynamic:    dynClient,
		Kubeconfig: kubeconfig,

		DatasourceConfig:     appconfig.NewDatasourceConfig(s.Datasource.Supported),
		DataserviceConnector: dw.NewConnector(s.Dataservice.MetadataHost, s.Dataservice.MetadataPort, s.Dataservice.DataHost, s.Dataservice.DataPort),

		TopicConfig:   appconfig.NewTopicConfig(s.Topic.Host, s.Topic.Port, s.Topic.HttpPort, s.Topic.AuthEnable, s.Topic.SuperUserToken, s.Topic.TokenSecret, s.Topic.PrestoHost, s.Topic.PrestoPort),
		Auditor:       audit.NewAuditor(s.Audit.Host, s.Audit.Port),
		TenantEnabled: s.TenantEnabled,
		SyncMode:      s.SyncMode,
		LocalConfig:   *errConfig,

		Mutex:   mtx,
		Elector: elector,

		Database: db,
	}
	casType := "cas"
	if s.TenantEnabled {
		casType = "tenant"
	} else {
		cache.SetGVs([]string{"apis", "applications", "applicationgroups", "applies", "datasources", "serviceunits", "serviceunitgroups", "topics", "topicgroups"})
	}
	cas.SetConnectionInfo(casType, s.Cas.Host, s.Cas.Port)
	return c, nil
}
