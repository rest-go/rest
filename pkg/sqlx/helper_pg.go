package sqlx

import "fmt"

type PGHelper struct{}

func (h PGHelper) GetTablesSQL() string {
	return `
	SELECT
		c.relname as name
	FROM
		pg_catalog.pg_class c
	LEFT JOIN
		pg_catalog.pg_namespace n ON n.oid = c.relnamespace
	WHERE
		c.relkind IN ('r','v','m','f','')
	  	AND n.nspname <> 'pg_catalog'
	  	AND n.nspname <> 'information_schema'
	  	AND n.nspname !~ '^pg_toast'
	  	AND pg_catalog.pg_table_is_visible(c.oid)
	ORDER BY 1
	`
}

func (h PGHelper) GetColumnsSQL(tableName string) string {
	return fmt.Sprintf(`
	SELECT column_name, data_type
	FROM information_schema.columns
	WHERE table_name = '%s' 
	ORDER BY ordinal_position;
	`, tableName)
}
