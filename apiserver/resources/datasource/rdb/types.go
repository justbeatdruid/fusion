package rdb

type Table struct {
	Name        string   `json:"tableName"`
	Type        string   `json:"tableType"`
	Tags        []string `json:"tags"`
	Description string   `json:"desc"`
}

type Field struct {
	Name        string `json:"name"`
	Unique      bool   `json:"unique"`
	DataType    string `json:"dataType"`
	Length      int    `json:"length"`
	IsNullAble  bool   `json:"IsNullAble"`
	Description string `json:"desc"`
	PrimaryKey  bool   `json:"primaryKey"`
	Index       int    `json:"index"`
}
