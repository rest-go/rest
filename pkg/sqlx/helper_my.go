package sqlx

import "fmt"

type MyHelper struct{}

func (h MyHelper) GetTablesSQL() string {
	return `
	SELECT TABLE_NAME as name
	FROM information_schema.TABLES
	WHERE TABLE_TYPE LIKE 'BASE_TABLE' AND TABLE_SCHEMA=DATABASE();
	`
}

func (h MyHelper) GetColumnsSQL(tableName string) string {
	return fmt.Sprintf(`
	SELECT COLUMN_NAME, DATA_TYPE
	FROM INFORMATION_SCHEMA. COLUMNS
	WHERE table_schema = DATABASE() AND table_name = '%s';
	`, tableName)
}
