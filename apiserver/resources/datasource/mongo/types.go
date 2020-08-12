package mongo

type Collection struct {
	Name string `json:"name"`
}

type Field struct {
	Name     string   `json:"name"`
	DataType DataType `json:"dataType"`
}

type DataType string

const (
	Integer DataType = "int"
	Float            = "float"
	Bool             = "bool"
	String           = "string"
	Unknown          = "unknown"
)
