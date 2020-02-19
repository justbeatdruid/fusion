package config

type DataserviceConfig struct {
	Host string
	Port int
}

func NewDataserviceConfig(host string, port int) *DataserviceConfig {
	return &DataserviceConfig{host, port}
}
