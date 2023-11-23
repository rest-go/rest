package sql

import "fmt"

type MyHelper struct{}

func (h MyHelper) GetTablesSQL() string {
	return `
	SELECT TABLE_NAME as name
	FROM information_schema.TABLES
	WHERE (TABLE_TYPE = 'BASE TABLE' OR TABLE_TYPE = 'view') AND TABLE_SCHEMA=DATABASE();
	`
}

func (h MyHelper) GetColumnsSQL(tableName string) string {
	return fmt.Sprintf(`
	SELECT
		COLUMN_NAME AS column_name,
		DATA_TYPE AS data_type,
		IS_NULLABLE="NO" AS notnull,
		COLUMN_KEY="PRI" AS pk
	FROM INFORMATION_SCHEMA.COLUMNS
	WHERE table_schema = DATABASE() AND table_name = '%s';
	`, tableName)
}
