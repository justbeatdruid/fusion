package datawarehouse

import (
	"encoding/json"
	"fmt"

	"github.com/chinamobile/nlpt/crds/datasource/datawarehouse/api/v1"
)

type Datawarehouse struct {
	Databases []Database `json:"data"`
}

type Database struct {
	Name string `json:"databaseName"`
}

type Table struct {
	Name        string   `json:"tableName"`
	Type        string   `json:"tableType"`
	Tags        []string `json:"tags"`
	Description string   `json:"desc"`
}

type Property struct {
	Name        string `json:"name"`
	Unique      bool   `json:"unique"`
	DataType    string `json:"dataType"`
	Length      int    `json:"length"`
	Description string `json:"desc"`
	PrimaryKey  bool   `json:"primaryKey"`
}

func FromApiTable(table v1.Table) Table {
	t := Table{}
	fromApi(&table, &t)
	return t
}

func FromApiProperty(property v1.Property) Property {
	p := Property{}
	fromApi(&property, &p)
	return p
}

func fromApi(api, model interface{}) error {
	b, err := json.Marshal(api)
	if err != nil {
		return fmt.Errorf("marshal error: %+v", err)
	}
	err = json.Unmarshal(b, model)
	if err != nil {
		return fmt.Errorf("unmarshal error: %+v", err)
	}
	return nil
}
