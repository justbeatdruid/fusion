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

	ConfigPath string

	Datasource  *DatasourceOptions
	Dataservice *DataserviceOptions
	Topic       *TopicOptions
	Cas         *CasOptions
	Audit       *AuditOptions
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		ListenAddress: ":8001",
		CrdNamespace:  os.Getenv("MY_POD_NAMESPACE"),
		TenantEnabled: false,
		ConfigPath:    "/data/err.json",
		Datasource:    DefaultDatasourceOptions(),
		Dataservice:   DefaultDataserviceOptions(),
		Topic:         DefaultTopicOptions(),
		Cas:           DefaultCasOptions(),
		Audit:         DefaultAuditOptions(),
	}

	if len(s.CrdNamespace) == 0 {
		klog.Infof("cannot find environmnent MY_POD_NAMESPACE, use default")
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
	kfset := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(kfset)
	fss.FlagSet("klog").AddGoFlagSet(kfset)
	fs := fss.FlagSet("misc")
	fs.StringVar(&s.Master, "master", s.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig).")
	fs.StringVar(&s.Kubeconfig, "kubeconfig", s.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")
	fs.StringVar(&s.ListenAddress, "listen", s.ListenAddress, "Address of app manager server listening on")
	fs.BoolVar(&s.TenantEnabled, "tenant-enabled", s.TenantEnabled, "Enable tenant, that means for on resource, role of user is always same for all items in tenant. If tenant enabled, we do not check anymore for user's writing or reading permission")
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

	errConfig, err := s.ParseLocalConfig()
	if err != nil {
		return nil, fmt.Errorf("parse error config error: %+v", err)
	}

	c := &appconfig.Config{
		Client:     client,
		Dynamic:    dynClient,
		Kubeconfig: kubeconfig,

		DatasourceConfig:     appconfig.NewDatasourceConfig(s.Datasource.Supported),
		DataserviceConnector: dw.NewConnector(s.Dataservice.MetadataHost, s.Dataservice.MetadataPort, s.Dataservice.DataHost, s.Dataservice.DataPort),

		TopicConfig:   appconfig.NewTopicConfig(s.Topic.Host, s.Topic.Port),
		Auditor:       audit.NewAuditor(s.Audit.Host, s.Audit.Port),
		TenantEnabled: s.TenantEnabled,
		LocalConfig:   *errConfig,
	}
	cas.SetConnectionInfo(s.Cas.Host, s.Cas.Port)
	return c, nil
}
