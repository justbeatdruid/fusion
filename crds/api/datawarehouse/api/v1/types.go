package v1

import (
	"encoding/json"
	"fmt"

	"github.com/chinamobile/nlpt/pkg/datawarehouse/api/v1"
	"github.com/chinamobile/nlpt/pkg/util"
)

type DataWarehouseQuery struct {
	Properties []QueryProperty `json:"properties"`
}

type QueryProperty struct {
	TableID      string `json:"tableId"`
	TableName    string `json:"tableName"`
	PropertyID   string `json:"propertyId"`
	PropertyName string `json:"propertyName"`
	PhysicalType string `json:"physicalType"`
}

func (d *DataWarehouseQuery) Validate() error {
	if d == nil {
		return nil
	}
	for _, p := range d.Properties {
		for k, v := range map[string]string{
			"table id":      p.TableID,
			"property id":   p.PropertyID,
			"table name":    p.TableName,
			"property name": p.PropertyName,
			"type":          p.PhysicalType,
		} {
			if len(v) == 0 {
				return fmt.Errorf("%s is null", k)
			}
		}
	}
	return nil
}

type Query struct {
	PrimaryTableName  string             `json:"primaryTableName"`
	AssociationTables []AssociationTable `json:"associationTable"`
	QueryFieldList    []QueryField       `json:"queryFieldList"`
	WhereFieldInfo    []WhereField       `json:"whereFieldInfo"`
	GroupbyFieldInfo  []GroupbyField     `json:"groupByFieldInfo"`
	Limit             int                `json:"limitNum"`
}

type AssociationTable struct {
	AssociationPropertyName string `json:"associationPropertyName"`
	AassociationTableName   string `json:"associationTableName"`
	PropertyName            string `json:"propertyName"`
	TableName               string `json:"tableName"`
}

type QueryField struct {
	PropertyName        string `json:"propertyName"`
	PropertyDisplayName string `json:"propertyDisplayName,omitempty"`
	TableName           string `json:"tableName"`
	Operator            string `json:"operator,omitempty"`
}

type WhereField struct {
	//DataType     string   `json:"dataType"`
	PropertyName string   `json:"propertyName"`
	TableName    string   `json:"tableName"`
	Operator     string   `json:"operator"`
	Values       []string `json:"value"`

	ParameterEnabled bool   `json:"parameterEnabled,omitempty"`
	Example          string `json:"example,omitempty"`
	Description      string `json:"description,omitempty"`
	Required         bool   `json:"required,omitempty"`
}

type GroupbyField struct {
	PropertyName string `json:"propertyName"`
	TableName    string `json:"tableName"`
}

func (w WhereField) ParamName() string {
	return fmt.Sprintf("%s/%s", w.TableName, w.PropertyName)
}

func (q QueryField) ParamName() string {
	return fmt.Sprintf("%s/%s", q.TableName, q.PropertyName)
}

func (q *DataWarehouseQuery) ToApiQuery(params map[string][]string) v1.Query {
	wheres := make([]v1.WhereField, 0)

	apiQuery := v1.Query{}
	toApi(q, &apiQuery)
	apiQuery.WhereFieldInfo = wheres
	return apiQuery
}

func (q *Query) Validate() error {
	if q == nil {
		return fmt.Errorf("query is null")
	}
	if len(q.PrimaryTableName) == 0 {
		return fmt.Errorf("primary table name is null")
	}
	for _, a := range q.AssociationTables {
		if err := util.CheckStruct(&a); err != nil {
			return fmt.Errorf("associationTable wrong: %+v", err)
		}
	}
	for _, a := range q.QueryFieldList {
		if err := util.CheckStruct(&a); err != nil {
			return fmt.Errorf("queryFieldList wrong: %+v", err)
		}
	}
	for _, a := range q.GroupbyFieldInfo {
		if err := util.CheckStruct(&a); err != nil {
			return fmt.Errorf("groupbyFieldInfo wrong: %+v", err)
		}
	}
	for _, w := range q.WhereFieldInfo {
		for k, v := range map[string]string{
			"propertyName": w.PropertyName,
			"tableName":    w.TableName,
		} {
			if len(v) == 0 {
				return fmt.Errorf("whereFieldList wrong: %s is null", k)
			}
		}
		if w.ParameterEnabled {
			for k, v := range map[string]string{
				"example":     w.Example,
				"description": w.Description,
			} {
				if len(v) == 0 {
					return fmt.Errorf("whereFieldList wrong: %s is null", k)
				}
			}
		}
	}
	return nil
}

func toApi(model, api interface{}) error {
	b, err := json.Marshal(model)
	if err != nil {
		return fmt.Errorf("marshal error: %+v", err)
	}
	err = json.Unmarshal(b, api)
	if err != nil {
		return fmt.Errorf("unmarshal error: %+v", err)
	}
	return nil
}
