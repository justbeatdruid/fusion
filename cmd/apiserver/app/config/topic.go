package config

type TopicConfig struct {
	Host       string
	Port       int
	AuthEnable bool
	AdminToken string
}

func NewTopicConfig(host string, port int, authEnable bool, adminToken string) *TopicConfig {
	return &TopicConfig{host, port, authEnable, adminToken}
}
