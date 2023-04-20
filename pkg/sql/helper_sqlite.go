package sql

import "fmt"

type SQLiteHelper struct{}

func (h SQLiteHelper) GetTablesSQL() string {
	return `
	SELECT 
    	name
	FROM 
    	sqlite_schema
	WHERE 
    	(type ='table' OR type = 'view') AND
    	name NOT LIKE 'sqlite_%';
	`
}

func (h SQLiteHelper) GetColumnsSQL(tableName string) string {
	return fmt.Sprintf(`
		SELECT 
			name as column_name,
			type as data_type,
			"notnull" = 1 as "notnull",
			pk >=1 as pk
		FROM PRAGMA_TABLE_INFO('%s')
	`, tableName)
}
