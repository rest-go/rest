package sql

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
	SELECT
		c.column_name,
		c.data_type,
		c.is_nullable='NO' as notnull,
		pc.contype='p' IS TRUE as pk
	FROM information_schema.columns c
	LEFT JOIN information_schema.key_column_usage kcu
	ON 
		c.column_name = kcu.column_name AND
		c.table_name = kcu.table_name
	LEFT JOIN pg_constraint pc
	ON 
		kcu.constraint_name=pc.conname
	WHERE c.table_name = '%s' 
	ORDER BY c.ordinal_position;
	`, tableName)
}
