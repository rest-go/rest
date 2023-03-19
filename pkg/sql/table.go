package sql

import (
	"fmt"
	"strings"
)

// Column represents a table column with name and type
type Column struct {
	ColumnName string `json:"column_name"`
	DataType   string `json:"data_type"`
	NotNull    bool   `json:"notnull"`
	Pk         bool   `json:"pk"`
}

func (c *Column) String() string {
	return fmt.Sprintf("%s %s", c.ColumnName, c.DataType)
}

// Table represents a table in database with name and columns
type Table struct {
	Name       string
	PrimaryKey string
	Columns    []*Column
}

func (t *Table) String() string {
	var columnsBuilder strings.Builder
	columnsBuilder.WriteString("(\n")
	for i, c := range t.Columns {
		columnsBuilder.WriteString("  ")
		columnsBuilder.WriteString(c.String())
		if i < len(t.Columns)-1 {
			columnsBuilder.WriteString(",\n")
		}
	}
	columnsBuilder.WriteString("\n)")
	return fmt.Sprintf("%s %s", t.Name, columnsBuilder.String())
}
