package config

type TopicConfig struct {
	Host string
	Port int
}

func NewTopicConfig(host string, port int) *TopicConfig {
	return &TopicConfig{host, port}
}
