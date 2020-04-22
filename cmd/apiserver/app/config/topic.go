package config

type TopicConfig struct {
	Host        string
	Port        int
	AuthEnable  bool
	AdminToken  string
	TokenSecret string
}

func NewTopicConfig(host string, port int, authEnable bool, adminToken string, tokenSecret string) *TopicConfig {
	return &TopicConfig{host, port, authEnable, adminToken, tokenSecret}
}
