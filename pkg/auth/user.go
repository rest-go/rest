package auth

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/rest-go/rest/pkg/log"
	"github.com/rest-go/rest/pkg/sql"
)

const (
	// The name of the users table
	UserTableName = "auth_users"

	createUserTable = `
	CREATE TABLE auth_users (
		id %s,
		username VARCHAR(32) UNIQUE NOT NULL,
		password VARCHAR(72) NOT NULL,
		is_admin bool NOT NULL DEFAULT false
	)
	`
	createAdminUser = `INSERT INTO auth_users (username, password, is_admin) VALUES (?, ?, true)`
	createUser      = `INSERT INTO auth_users (username, password) VALUES (?, ?)`
	queryUser       = `SELECT id, username, password, is_admin FROM auth_users WHERE username = ?`
)

// User represents a request user
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

// IsAuthenticated returns a bool to indicate whether user is anonymous
func (u *User) IsAnonymous() bool {
	return u.ID == 0
}

// IsAuthenticated returns a bool to indicate whether user is authenticated
func (u *User) IsAuthenticated() bool {
	return u.ID != 0
}

func (u *User) hasPerm(exp string) (hasPerm bool, withUserIDColumn string) {
	// remove all the spaces in expression
	exp = strings.ReplaceAll(exp, " ", "")
	// if ask a admin user perm
	if exp == "" {
		return true, ""
	} else if exp == "auth_user.is_admin" {
		return u.IsAdmin, ""
	} else if exp == "auth_user.is_authenticated" {
		return u.IsAuthenticated(), ""
	} else if strings.HasSuffix(exp, "=auth_user.id") {
		return u.IsAuthenticated(), strings.TrimSuffix(exp, "=auth_user.id")
	}

	log.Errorf("invalid policy exp: %s, return false", exp)
	return false, ""
}

// HasPerm check whether user has permission to perform action on the table with provided policies
func (u *User) HasPerm(table string, action Action, policies map[string]map[string]string) (hasPerm bool, withUserIDColumn string) {
	if policies == nil {
		log.Warnf("nil policies")
		return false, ""
	}

	var ps map[string]string
	ps, ok := policies[table]
	defaultTablePerm := policies["all"]
	if !ok {
		ps = defaultTablePerm
	}
	if len(ps) > 0 {
		if exp, ok := ps[action.String()]; ok {
			return u.hasPerm(exp)
		} else if exp, ok := ps["all"]; ok {
			return u.hasPerm(exp)
		} else {
			return u.hasPerm(defaultTablePerm["all"])
		}
	}

	return true, ""
}

// HashPassword generate the hashed password for a plain password
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("generate hashed password error %w", err)
	}
	return string(hashedPassword), nil
}

func genPasswd(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(randomBytes)[:length], nil
}

// setupUsers create `users` table and create an admin user
func setupUsers(db *sql.DB) (username, password string, err error) {
	log.Info("create users table")
	idSQL := primaryKeySQL[db.DriverName]
	createTableQuery := fmt.Sprintf(createUserTable, idSQL)
	ctx, cancel := context.WithTimeout(context.Background(), sql.DefaultTimeout)
	defer cancel()
	_, dbErr := db.ExecQuery(ctx, createTableQuery)
	if dbErr != nil {
		return "", "", dbErr
	}

	log.Info("create a admin user")
	username = adminUsername
	length := 12
	password, err = genPasswd(length)
	if err != nil {
		return "", "", err
	}
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return "", "", err
	}
	_, dbErr = db.ExecQuery(ctx, createAdminUser, username, hashedPassword)
	return username, password, dbErr
}
