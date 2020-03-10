package v1

import (
	"fmt"
	"strings"

	crdv1 "github.com/chinamobile/nlpt/crds/datasource/datawarehouse/api/v1"
	"github.com/chinamobile/nlpt/pkg/datawarehouse/api/v1"

	"k8s.io/klog"
)

type DataWarehouseQuery struct {
	Properties     []QueryProperty `json:"properties"`
	PrimaryTableID string          `json:"primaryTableId"`
	Query          *v1.Query       `json:"query"`
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
	if len(d.PrimaryTableID) == 0 {
		return fmt.Errorf("primary table id not set")
	}
	if len(d.Properties) == 0 {
		return fmt.Errorf("cannot find properties")
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

func NewQuery() *v1.Query {
	return &v1.Query{
		UserID: "admin",
	}
}

type TableNode struct {
	ID               string
	Name             string
	ParentTableID    string
	ParentTableName  string
	ParentFieldID    string
	ParentFieldName  string
	RelatedFieldID   string
	RelatedFieldName string
	IsRoot           bool
	SubTables        []*TableNode
}

func (dq *DataWarehouseQuery) RefillQuery(db *crdv1.Database) error {
	if dq == nil {
		return fmt.Errorf("data warehouse query is null")
	}
	if db == nil {
		return fmt.Errorf("database is null")
	}
	dq.Query = NewQuery()

	// Step 1: build basic info
	dq.Query.DatabaseName = db.Name
	var primaryTable crdv1.Table
	for _, t := range db.Tables {
		if t.Info.ID == dq.PrimaryTableID {
			dq.Query.PrimaryTableName = t.Info.Name
			primaryTable = t
		}
	}
	if len(dq.Query.PrimaryTableName) == 0 {
		return fmt.Errorf("cannot find table with ID %s in database(ID) %s", dq.PrimaryTableID, db.Id)
	}

	// Step 2: build association tables
	// build a tree firstly
	// IMPORTANT v1.Table should never be used
	//           we use crdv1.Table to build v1.Query
	{
		selectedTables := make(map[string]crdv1.Table)
		contains := func(t crdv1.Table) bool {
			for _, p := range dq.Properties {
				if p.TableID == t.Info.ID {
					return true
				}
			}
			return false
		}
		for _, t := range db.Tables {
			if contains(t) {
				selectedTables[t.Info.ID] = t
			}
		}
		tnroot := &TableNode{
			ID:     primaryTable.Info.ID,
			Name:   primaryTable.Info.Name,
			IsRoot: true,
		}
		var addChildren func(tn *TableNode)
		addChildren = func(tn *TableNode) {
			tn.SubTables = make([]*TableNode, 0)
			parentTable := selectedTables[tn.ID]
			for _, parentField := range parentTable.Properties {
				if len(parentField.ReferenceTableId) > 0 && len(parentField.ReferencePropertyId) > 0 {
					if childTable, ok := selectedTables[parentField.ReferenceTableId]; ok {
						getFieldByID := func(fieldID string) crdv1.Property {
							for _, f := range childTable.Properties {
								if f.ID == fieldID {
									return f
								}
							}
							klog.Errorf("cannot find (in database %s) referenced field by ID: parent table id=%s, parent field id=%s, "+
								"parent shows that referenced table id=%s, referenced field id=%s, but the field not found in child table",
								db.Id, parentTable.Info.ID, parentField.ID, parentField.ReferenceTableId, parentField.ReferencePropertyId)
							return crdv1.Property{}
						}
						tn.SubTables = append(tn.SubTables, &TableNode{
							ID:               childTable.Info.ID,
							Name:             childTable.Info.Name,
							ParentTableID:    parentTable.Info.ID,
							ParentTableName:  parentTable.Info.Name,
							ParentFieldID:    parentField.ID,
							ParentFieldName:  parentField.Name,
							RelatedFieldID:   parentField.ReferencePropertyId,
							RelatedFieldName: getFieldByID(parentField.ReferencePropertyId).Name,
							IsRoot:           false,
						})
					}
				}
			}
			for _, subnode := range tn.SubTables {
				addChildren(subnode)
			}
		}
		addChildren(tnroot)
		dq.Query.AssociationTables = make([]v1.AssociationTable, 0)
		var addAssociations func(tn *TableNode)
		addAssociations = func(tn *TableNode) {
			if tn == nil {
				klog.Errorf("it seems there is a null table node")
				return
			}
			if !tn.IsRoot {
				dq.Query.AssociationTables = append(dq.Query.AssociationTables, v1.AssociationTable{
					AssociationPropertyName: tn.ParentFieldName,
					AassociationTableName:   tn.ParentTableName,
					PropertyName:            tn.RelatedFieldName,
					TableName:               tn.Name,
				})
			}
			for _, subnode := range tn.SubTables {
				addAssociations(subnode)
			}
		}
		addAssociations(tnroot)
	}

	// Step 3: build query field list
	{
		dq.Query.QueryFieldList = make([]v1.QueryField, 0)
		getFieldInfo := func(tableID, fieldID string) (string, string, string) {
			for _, t := range db.Tables {
				if t.Info.ID == tableID {
					for _, p := range t.Properties {
						if p.ID == fieldID {
							return t.Info.Name, p.Name, p.DisplayName
						}
					}
				}
			}
			klog.Errorf("cannot find table with ID %s and property with ID %s in database with ID %s",
				tableID, fieldID, db.Id)
			return "", "", ""
		}
		for _, p := range dq.Properties {
			tn, pn, pdn := getFieldInfo(p.TableID, p.PropertyID)
			dq.Query.QueryFieldList = append(dq.Query.QueryFieldList, v1.QueryField{
				TableName:           tn,
				PropertyName:        pn,
				PropertyDisplayName: pdn,
			})
		}
	}

	// Step 4: don't leave other fields null
	// Where fileds will be built when query with query params
	dq.Query.WhereFieldInfo = make([]v1.WhereField, 0)
	dq.Query.GroupbyFieldInfo = make([]v1.GroupbyField, 0)
	return nil
}

func (dq *DataWarehouseQuery) RefillWhereFields(typesMap map[string]string, params map[string][]string) error {
	dq.Query.WhereFieldInfo = make([]v1.WhereField, 0)
	for pk, pv := range params {
		ss := strings.Split(pk, ".")
		if len(ss) != 2 {
			return fmt.Errorf("Query parameter %s is wrong. expect format is {tableName}.{propertyName}", pk)
		}
		if tv, ok := typesMap[pk]; ok {
			dq.Query.WhereFieldInfo = append(dq.Query.WhereFieldInfo, v1.WhereField{
				TableName:    ss[0],
				PropertyName: ss[1],
				Operator:     "equal",
				DataType:     tv,
				Values:       pv,
			})
		}
	}
	return nil
}
