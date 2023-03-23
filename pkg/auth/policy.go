package auth

import (
	"context"
	"fmt"

	"github.com/rest-go/rest/pkg/log"
	"github.com/rest-go/rest/pkg/sql"
)

const (
	// the name of the policies table
	PolicyTableName = "auth_policies"

	createPolicyTable = `
	CREATE TABLE auth_policies (
		id %s,
		description VARCHAR(256) NOT NULL,
		table_name VARCHAR(128) NOT NULL,
		action VARCHAR(16) NOT NULL,
		expression VARCHAR(128) NOT NULL
	)
	`
	createInternalPolicy = `
		INSERT INTO auth_policies (description, table_name, action, expression)
		VALUES (?, ?, ?, ?)
	`
)

var defaultPolicies = []Policy{
	{
		Description: "policies operations are limited to admin user",
		TableName:   "auth_policies",
		Action:      "all",
		Expression:  "auth_user.is_admin",
	},
	{
		Description: "users are limited to admin user(to deny user to update self to admin)",
		TableName:   "auth_users",
		Action:      "all",
		Expression:  "auth_user.is_admin",
	},
	{
		Description: "all tables/actions are limited to be filtered by user_id",
		TableName:   "all",
		Action:      "all",
		Expression:  "user_id = auth_user.id",
	},
}

// Policy represents a security policy against a table
type Policy struct {
	ID          int64  `json:"id"`
	Description string `json:"description"`
	TableName   string `json:"table_name"`
	Action      string `json:"action"`
	Expression  string `json:"expression"`
}

// setupPolicies create `policies` table and create a default internal policies
func setupPolicies(db *sql.DB) error {
	log.Info("create policies table")
	idSQL := primaryKeySQL[db.DriverName]
	createTableQuery := fmt.Sprintf(createPolicyTable, idSQL)
	ctx, cancel := context.WithTimeout(context.Background(), sql.DefaultTimeout)
	defer cancel()
	_, dbErr := db.ExecQuery(ctx, createTableQuery)
	if dbErr != nil {
		return dbErr
	}

	log.Info("create default policies")
	for _, policy := range defaultPolicies {
		_, dbErr := db.ExecQuery(
			ctx,
			createInternalPolicy,
			policy.Description,
			policy.TableName,
			policy.Action,
			policy.Expression,
		)
		if dbErr != nil {
			return dbErr
		}
	}
	return nil
}
