package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var policies = map[string]map[string]string{
	// Default policies
	// users are limited to admin user
	"users": {
		"all": "auth_user.is_admin",
	},
	// policies operations are limited to admin user
	"policies": {
		"all": "auth_user.is_admin",
	},
	// all tables are limited to filter by user_id by default
	"all": {
		"all": "user_id = auth_user.id",
	},

	// Custom policies
	// todos operations are limited by `author_id` field
	"todos": {
		"all": "author_id = auth_user.id",
	},
	"articles": {
		"read": "",
		"all":  "author_id = auth_user.id",
	},
	"comments": {
		"read": "",
	},
	"reviews": {
		"read": "auth_user.is_authenticated",
	},
	"notes": {
		"read": "invalid policy",
	},
}

//nolint:funlen
func TestUser_HasPerm(t *testing.T) {
	for _, test := range []struct {
		name             string
		user             User
		table            string
		action           Action
		hasPerm          bool
		withUserIDColumn string
	}{
		{
			name:             "non-admin users don't have permission on users table",
			user:             User{IsAdmin: false, ID: 1},
			table:            "users",
			action:           ActionRead,
			hasPerm:          false,
			withUserIDColumn: "",
		},
		{
			name:             "admin users have permission on users tables",
			user:             User{IsAdmin: true, ID: 1},
			table:            "users",
			action:           ActionUpdate,
			hasPerm:          true,
			withUserIDColumn: "",
		},
		{
			name:             "non-admin users don't have permission on policies",
			user:             User{IsAdmin: false, ID: 1},
			table:            "policies",
			action:           ActionRead,
			hasPerm:          false,
			withUserIDColumn: "",
		},
		{
			name:             "admin users have permission on policies",
			user:             User{IsAdmin: true, ID: 1},
			table:            "policies",
			action:           ActionRead,
			hasPerm:          true,
			withUserIDColumn: "",
		},
		{
			name:             "limit by `user_id` field by default",
			user:             User{IsAdmin: false, ID: 1},
			table:            "some_table",
			action:           ActionRead,
			hasPerm:          true,
			withUserIDColumn: "user_id",
		},
		{
			name:             "users have read permission on todos with custom column",
			user:             User{IsAdmin: false, ID: 1},
			table:            "todos",
			action:           ActionRead,
			hasPerm:          true,
			withUserIDColumn: "author_id",
		},
		{
			name:             "users have write permission on todos with custom column",
			user:             User{IsAdmin: false, ID: 1},
			table:            "todos",
			action:           ActionCreate,
			hasPerm:          true,
			withUserIDColumn: "author_id",
		},
		{
			name:             "users have permission on read articles",
			user:             User{IsAdmin: false, ID: 1},
			table:            "articles",
			action:           ActionRead,
			hasPerm:          true,
			withUserIDColumn: "",
		},
		{
			name:             "limit to current auth user when read mine",
			user:             User{IsAdmin: false, ID: 1},
			table:            "articles",
			action:           ActionReadMine,
			hasPerm:          true,
			withUserIDColumn: "author_id",
		},
		{
			name:             "comments has public read perm",
			user:             User{IsAdmin: false, ID: 1},
			table:            "comments",
			action:           ActionRead,
			hasPerm:          true,
			withUserIDColumn: "",
		},
		{
			name:             "comments has default other perm",
			user:             User{ID: 0},
			table:            "comments",
			action:           ActionCreate,
			hasPerm:          false,
			withUserIDColumn: "user_id",
		},
		{
			name:             "reviews has public read perm for authenticated user",
			user:             User{ID: 1},
			table:            "reviews",
			action:           ActionRead,
			hasPerm:          true,
			withUserIDColumn: "",
		},
		{
			name:             "reviews has not allow read perm for anonymous user",
			user:             User{ID: 0},
			table:            "reviews",
			action:           ActionRead,
			hasPerm:          false,
			withUserIDColumn: "",
		},
		{
			name:             "notes has invalid policy, return false",
			user:             User{ID: 1, IsAdmin: true},
			table:            "notes",
			action:           ActionRead,
			hasPerm:          false,
			withUserIDColumn: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			hasPerm, userIDColumn := test.user.HasPerm(test.table, test.action, policies)
			assert.Equal(t, test.hasPerm, hasPerm)
			assert.Equal(t, test.withUserIDColumn, userIDColumn)
		})
	}
	t.Run("nil policies, return false", func(t *testing.T) {
		user := User{ID: 1, IsAdmin: true}
		hasPerm, userIDColumn := user.HasPerm("users", ActionRead, nil)
		assert.False(t, hasPerm)
		assert.Equal(t, "", userIDColumn)
	})
}
