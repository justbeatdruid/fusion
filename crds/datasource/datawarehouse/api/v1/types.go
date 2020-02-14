package v1

type Datawarehouse struct {
	Databases []Database `json:"data"`
}

type Database struct {
	Name   string  `json:"databaseName"`
	Tables []Table `json:"table_property"`
}

type Table struct {
	Name        string     `json:"tableName"`
	Type        string     `json:"tableType"`
	Tags        []string   `json:"tags"`
	Description string     `json:"desc"`
	Properties  []Property `json:"property"`
}

type Property struct {
	TableID         string `json:"tableId"`
	ID              int    `json:"id"`
	Name            string `json:"name"`
	DisplayName     string `json:"displayName"`
	Unique          string `json:"unique"`
	DataType        string `json:"dataType"`
	Length          int    `json:"length"`
	Description     string `json:"desc"`
	Encryption      string `json:"encryption"`
	EncrypAlgorithm string `json:"encrypAlgorithm"`
	PrimaryKey      string `json:"primaryKey"`
}

type Query struct {
	// always admin
	UserID string `json:"userId"`

	PrimaryTableName  string             `json:"primaryTableName"`
	AssociationTables []AssociationTable `json:"associationTable"`
	QueryFieldList    []QueryField       `json:"queryFieldList"`
	WhereFieldInfo    []WhereField       `json:"whereFieldInfo"`
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
	DataType     string   `json:"dataType"`
	PropertyName string   `json:"propertyName"`
	TableName    string   `json:"tableName"`
	Values       []string `json:"value"`
	Operator     string   `json:"operator"`
}

type Result struct {
	Headers   []string            `json:"headerList"`
	ColumnDic map[string]string   `json:"columnDic"`
	Data      []map[string]string `json:"dataValueList"`
}
