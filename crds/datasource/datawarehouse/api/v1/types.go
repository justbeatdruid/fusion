package v1

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/chinamobile/nlpt/pkg/datawarehouse/api/v1"
)

type Database struct {
	Id                 string  `json:"databaseId"`
	Name               string  `json:"databaseName"`
	DisplayName        string  `json:"databaseDisplayName"`
	SubjectId          string  `json:"subjectId"`
	SubjectName        string  `json:"subjectName"`
	SubjectDisplayName string  `json:"subjectDisplayName"`
	Tables             []Table `json:"tableMetadataInfos,omitempty"`
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

func (db *Database) GetTables(associationID string) (ts []Table) {
	if db == nil {
		return
	}
	if len(associationID) == 0 {
		for _, t := range db.Tables {
			if t.Info.Type == "事实表" {
				ts = append(ts, t)
			}
		}
		return
	} else {
		for _, t := range db.Tables {
			if t.Info.ID == associationID {
				return db.GetRelatedTables(t)
			}
		}
	}
	return
}

func (db *Database) GetRelatedTables(t Table) (ts []Table) {
	if db == nil {
		return
	}
	for _, p := range t.Properties {
		if len(p.ReferenceTableId) > 0 && len(p.ReferencePropertyId) > 0 {
			for _, t := range db.Tables {
				if t.Info.ID == p.ReferenceTableId {
					ts = append(ts, t)
				}
			}
		}
	}
	return
}

func FromApiDatabase(db v1.Database) Database {
	d := Database{}
	fromApi(&db, &d)
	return d
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

func OnlyTable(t Table) Table {
	t.Properties = nil
	return t
}

type TableList []Table

func (tl TableList) Len() int {
	return len(tl)
}

func (tl TableList) Less(i, j int) bool {
	return tl[i].Info.ID < tl[j].Info.ID
}

func (tl TableList) Swap(i, j int) {
	tl[i], tl[j] = tl[j], tl[i]
}

func DeepCompareDataWarehouse(d1, d2 *Database) (bool, error) { //d1 local  d2 orgen
	if d1.Tables == nil {
		d1.Tables = make([]Table, 0)
	}
	if d2.Tables == nil {
		d2.Tables = make([]Table, 0)
	}
	var ts1 TableList = d1.Tables
	var ts2 TableList = d2.Tables
	if ts1.Len() != ts2.Len() {
		return false, nil
	}
	if len(d1.Tables) == 0 {
		return true, nil
	}

	sort.Sort(ts1)
	sort.Sort(ts2)

	l := ts1.Len()

	for i := 0; i < l; i++ {
		if ts1[i].Info.ID != ts2[i].Info.ID {
			return false, nil
		}
		if ts1[i].Info.LastUpdateTime != ts2[i].Info.LastUpdateTime {
			return false, nil
		}
	}

	return true, nil
}
