package v1

import (
	"fmt"

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

	ParameterEnabled bool   `json:"parameterEnabled"`
	Type             string `json:"type"`
	Example          string `json:"example"`
	Description      string `json:"description"`
	Required         bool   `json:"required"`
}

type GroupbyField struct {
	PropertyName string `json:"propertyName"`
	TableName    string `json:"tableName"`
}

func (q *Query) Validate() error {
	if q == nil {
		return fmt.Errorf("query is null")
	}
	if len(q.PrimaryTableName) == 0 {
		return fmt.Errorf("primary table name is null")
	}
	if err := util.CheckStruct(q.AssociationTables); err != nil {
		return fmt.Errorf("associationTable wrong: %+v", err)
	}
	if err := util.CheckStruct(q.QueryFieldList); err != nil {
		return fmt.Errorf("queryFieldList wrong: %+v", err)
	}
	return nil
}
