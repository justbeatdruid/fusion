package v1

import (
	"encoding/json"
	"fmt"

	"github.com/chinamobile/nlpt/pkg/datawarehouse/api/v1"
	"github.com/chinamobile/nlpt/pkg/util"
)

type Query struct {
	// always admin
	UserID string `json:"userId"`

	DatabaseName      string             `json:"databaseName"`
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
	PropertyDisplayName string `json:"propertyDisplayName"`
	TableName           string `json:"tableName"`
	Operator            string `json:"operator"`
}

type WhereField struct {
	//DataType     string   `json:"dataType"`
	PropertyName string   `json:"propertyName"`
	TableName    string   `json:"tableName"`
	Operator     string   `json:"operator"`
	Values       []string `json:"value"`

	ParameterEnabled bool   `json:"parameterEnabled,omitempty"`
	Type             string `json:"type,omitempty"`
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

func (q *Query) ToApiQuery(params map[string][]string) v1.Query {
	wheres := make([]v1.WhereField, 0)
	for _, w := range q.WhereFieldInfo {
		apiWhere := v1.WhereField{
			PropertyName: w.PropertyName,
			TableName:    w.TableName,
			Operator:     w.Operator,
			Values:       w.Values,
		}
		if !w.ParameterEnabled {
			wheres = append(wheres, apiWhere)
		} else {
			for k, v := range params {
				if k == w.ParamName() {
					apiWhere.Values = v
					wheres = append(wheres, apiWhere)
				}
			}
		}
	}

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
		if err := util.CheckStruct(a); err != nil {
			return fmt.Errorf("associationTable wrong: %+v", err)
		}
	}
	for _, a := range q.QueryFieldList {
		if err := util.CheckStruct(a); err != nil {
			return fmt.Errorf("queryFieldList wrong: %+v", err)
		}
	}
	for _, a := range q.GroupbyFieldInfo {
		if err := util.CheckStruct(a); err != nil {
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
