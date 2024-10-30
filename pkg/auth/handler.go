package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	j "github.com/rest-go/rest/pkg/jsonutil"
	"github.com/rest-go/rest/pkg/log"
	"github.com/rest-go/rest/pkg/sql"
	"golang.org/x/crypto/bcrypt"
)

const adminUsername = "rest_admin"

// Handler is handler with auth endpoints like `register`, `login`, and `logout`
type Handler struct {
	db     *sql.DB
	secret []byte
}

// NewHandler return a Handler with provided database url and JWT secret
func NewHandler(dbURL string, secret []byte) (*Handler, error) {
	db, err := sql.Open(dbURL)
	if err != nil {
		return nil, err
	}
	return &Handler{db, secret}, nil
}

// ServeHTTP implements http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		res := &j.Response{
			Code: http.StatusMethodNotAllowed,
			Msg:  fmt.Sprintf("method not supported: %s", r.Method),
		}
		j.Write(w, res)
		return
	}

	action := strings.TrimPrefix(r.URL.Path, "/auth/")
	if action == "" {
		res := &j.Response{
			Code: http.StatusBadRequest,
			Msg:  "no auth action provided",
		}
		j.Write(w, res)
		return
	}

	var res any
	switch action {
	case "setup":
		res = h.setup()
	case "register":
		res = h.register(r)
	case "login":
		res = h.login(r)
	case "logout":
		res = h.logout(r)
	default:
		res = &j.Response{
			Code: http.StatusBadRequest,
			Msg:  "action not supported",
		}
	}
	j.Write(w, res)
}

func (h *Handler) setup() any {
	username, password, err := Setup(h.db)
	if err != nil {
		log.Error("setup error: ", err)
		return j.ErrResponse(err)
	}

	return &struct {
		Username string
		Password string
	}{
		Username: username,
		Password: password,
	}
}

func (h *Handler) register(r *http.Request) any {
	user := &User{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return &j.Response{
			Code: http.StatusBadRequest,
			Msg:  "failed to decode json data",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), sql.DefaultTimeout)
	defer cancel()
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return &j.Response{
			Code: http.StatusInternalServerError,
			Msg:  "failed to hash password",
		}
	}
	_, dbErr := h.db.ExecQuery(ctx, createUser, user.Username, hashedPassword)
	if dbErr != nil {
		log.Errorf("create user error: %v", dbErr)
		return j.ErrResponse(dbErr)
	}

	return &j.Response{Code: http.StatusOK, Msg: "success"}
}

func (h *Handler) login(r *http.Request) any {
	user := &User{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Warnf("failed to parse json data: %v", err)
		return &j.Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to parse post json data, %v", err),
		}
	}

	// authenticate the user by input username and password
	user, err = h.authenticate(user.Username, user.Password)
	if err != nil {
		log.Errorf("authenticate user error: %v", err)
		var dbErr sql.Error
		if errors.As(err, &dbErr) {
			return j.ErrResponse(dbErr)
		} else {
			return &j.Response{
				Code: http.StatusUnauthorized,
				Msg:  fmt.Sprintf("failed to authenticate user, %v", err),
			}
		}
	}

	tokenString, err := GenJWTToken(h.secret, map[string]any{
		"user_id":  user.ID,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(14 * 24 * time.Hour).Unix(),
	})
	if err != nil {
		return &j.Response{
			Code: http.StatusBadRequest,
			Msg:  fmt.Sprintf("failed to generate token, %v", err),
		}
	}

	return &struct {
		Token string `json:"token"`
	}{tokenString}
}

func (h *Handler) logout(_ *http.Request) any {
	// client delete token, no op on server side
	return &j.Response{Code: http.StatusOK, Msg: "success"}
}

func (h *Handler) authenticate(username, password string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sql.DefaultTimeout)
	defer cancel()

	row, dbErr := h.db.FetchOne(ctx, queryUser, username)
	if dbErr != nil {
		log.Errorf("fetch user error: %v", dbErr)
		return nil, dbErr
	}
	hashedPassword := row["password"].(string)
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return nil, errors.New("password doesn't match")
	}
	user := &User{
		ID:       row["id"].(int64),
		Username: username,
		IsAdmin:  row["is_admin"].(bool),
	}
	return user, nil
}
