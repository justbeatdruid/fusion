package config

type TopicConfig struct {
	Host        string
	Port        int
	AuthEnable  bool
	AdminToken  string
	TokenSecret string
	PrestoHost  string
	PrestoPort  int
}

func NewTopicConfig(host string, port int, authEnable bool, adminToken string, tokenSecret string, prestoHost string, prestoPort int) *TopicConfig {
	return &TopicConfig{host, port, authEnable, adminToken, tokenSecret, prestoHost, prestoPort}
}
