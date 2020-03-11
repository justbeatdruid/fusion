package config

import (
	dw "github.com/chinamobile/nlpt/pkg/datawarehouse"
)

type DataserviceConfig struct {
	Connector dw.Connector
}

func NewDataserviceConfig(metadatahost string, metadataport int, datahost string, dataport int) *DataserviceConfig {
	return &DataserviceConfig{
		Connector: dw.NewConnector(metadatahost, metadataport, datahost, dataport),
	}
}
