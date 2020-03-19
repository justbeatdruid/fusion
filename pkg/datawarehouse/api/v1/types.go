package v1

//types for quering dataservice

type Datawarehouse struct {
	Databases []Database `json:"data"`
}

type Database struct {
	Id                 string  `json:"databaseId"`
	Name               string  `json:"databaseName"`
	DisplayName        string  `json:"databaseDisplayName"`
	SubjectId          string  `json:"subjectId"`
	SubjectName        string  `json:"subjectName"`
	SubjectDisplayName string  `json:"subjectDisplayName"`
	Tables             []Table `json:"tableMetadataInfos"`
}

type Table struct {
	Info       TableInfo  `json:"tableInfo"`
	Properties []Property `json:"propertyEntrys"`
}

type TableInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	DisplayName    string `json:"displayName"`
	Type           string `json:"tableType"`
	EnglishName    string `json:"englishName"`
	CreateTime     string `json:"createTime"`
	LastUpdateTime string `json:"lastUpdateTime"`
	Schema         string `json:"schama"`
}

//{"id":"2928","name":"doctor_pk","displayName":"医生pk","englishName":null,"tableType":null,"tableId":null,"physicalType":"varchar","logicalType":"通用字符串","idx":0,"fieldLength":"0","fieldPersion":null,"isUnique":"否","des":null,"isPrimarykey":"否","isForeignkey":"是","referenceTableId":"2b7efd2a859f47da98ef5be248097a3a","referenceTableDisplayName":"医生信息维度表","referencePropertyId":"2698","referencePropertyName":"doctor_pk","isEncryption":"\u0000","entryptionType":null,"version":0,"standard":null,"isPartionfield":null,"sourceSql":null,"sourceTableId":null,"sourcePropertyId":null,"encrypt":"不加密"}
type Property struct {
	ID                        string `json:"id"`
	Name                      string `json:"name"`
	DisplayName               string `json:"displayName"`
	EnglishName               string `json:"englishName"`
	TableType                 string `json:"tableType"`
	TableId                   string `json:"tableId"`
	PhysicalType              string `json:"physicalType"`
	LogicalType               string `json:"logicalType"`
	Idx                       int    `json:"idx"`
	FieldLength               string `json:"fieldLength"`
	FieldPersion              string `json:"fieldPersion"`
	IsUnique                  string `json:"isUnique"`
	Des                       string `json:"des"`
	IsPrimarykey              string `json:"isPrimarykey"`
	IsForeignkey              string `json:"isForeignkey"`
	ReferenceTableId          string `json:"referenceTableId"`
	ReferenceTableDisplayName string `json:"referenceTableDisplayName"`
	ReferencePropertyId       string `json:"referencePropertyId"`
	ReferencePropertyName     string `json:"referencePropertyName"`
	IsEncryption              string `json:"isEncryption"`
	EntryptionType            string `json:"entryptionType"`
	Version                   int    `json:"version"`
	Standard                  string `json:"standard"`
	IsPartionfield            string `json:"isPartionfield"`
	SourceSql                 string `json:"sourceSql"`
	SourceTableId             string `json:"sourceTableId"`
	SourcePropertyId          string `json:"sourcePropertyId"`
	Encrypt                   string `json:"encrypt"`
}

type Query struct {
	// always admin
	UserID string `json:"userId"`

	DatabaseName      string             `json:"databaseName"`
	PrimaryTableName  string             `json:"primaryTableName"`
	AssociationTables []AssociationTable `json:"associationTable"`
	QueryFieldList    []QueryField       `json:"queryFieldList"`
	WhereFieldInfo    []WhereField       `json:"whereFieldInfo"`
	GroupbyFieldInfo  []GroupbyField     `json:"groupByFieldInfo"`
	Limit             int                `json:"limitNum,omitempty"`
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
	Operator            string `json:"queryOperator,omitempty"`
}

type WhereField struct {
	//DataType     string   `json:"dataType"`
	PropertyName string   `json:"propertyName"`
	TableName    string   `json:"tableName"`
	Operator     string   `json:"whereOperator"`
	DataType     string   `json:"dataType"`
	Values       []string `json:"value"`
}

type GroupbyField struct {
	PropertyName string `json:"propertyName"`
	TableName    string `json:"tableName"`
}

type Result struct {
	Headers   []string            `json:"headerList"`
	ColumnDic map[string]string   `json:"columnDic"`
	Data      []map[string]string `json:"dataValueList"`
}
