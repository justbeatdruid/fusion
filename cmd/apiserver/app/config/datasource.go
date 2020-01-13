package config

import (
	"strings"
)

type DatasourceConfig struct {
	Supported []string
}

func NewDatasourceConfig(supported string) *DatasourceConfig {
	return &DatasourceConfig{
		Supported: strings.Split(supported, ","),
	}
}
