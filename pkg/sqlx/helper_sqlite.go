package sqlx

import "fmt"

type SQLiteHelper struct{}

func (h SQLiteHelper) GetTablesSQL() string {
	return `
	SELECT 
    	name
	FROM 
    	sqlite_schema
	WHERE 
    	type ='table' AND 
    	name NOT LIKE 'sqlite_%';
	`
}

func (h SQLiteHelper) GetColumnsSQL(tableName string) string {
	return fmt.Sprintf(`
		SELECT name as column_name, type as data_type
		FROM PRAGMA_TABLE_INFO('%s')
	`, tableName)
}
