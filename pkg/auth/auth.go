// package auth provide restful interface for authentication
package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rest-go/rest/pkg/sql"
)

var primaryKeySQL = map[string]string{
	"postgres": "BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY",
	"mysql":    "BIGINT PRIMARY KEY AUTO_INCREMENT",
	"sqlite":   "INTEGER PRIMARY KEY",
}

// GenJWTToken generate and return jwt token
func GenJWTToken(secret []byte, data map[string]any) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(data))
	return token.SignedString(secret)
}

// ParseJWTToken parse tokenString and return data if token is valid
func ParseJWTToken(secret []byte, tokenString string) (map[string]any, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return map[string]any(claims), nil
	}

	return nil, errors.New("invalid token")
}

// Setup setup database tables and create an admin user account
func Setup(db *sql.DB) (username, password string, err error) {
	if isSetupDone(db) {
		err = errors.New("setup is already done before")
		return
	}
	username, password, err = setupUsers(db)
	if err != nil {
		return
	}
	err = setupPolicies(db)
	return
}

func isSetupDone(db *sql.DB) bool {
	ctx, cancel := context.WithTimeout(context.Background(), sql.DefaultTimeout)
	defer cancel()
	_, err := db.ExecQuery(ctx, "SELECT 1 FROM auth_users")
	return err == nil
}
