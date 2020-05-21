package config

type DatabaseConfig struct {
	Enabled  bool
	Type     string
	Host     string
	Port     int
	Username string
	Password string
	Database string
	Schema   string
}

func NewDatabaseConfig(e bool, t, h string, p int, u, s, d, c string) *DatabaseConfig {
	return &DatabaseConfig{e, t, h, p, u, s, d, c}
}
