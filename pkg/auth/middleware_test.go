package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	j "github.com/rest-go/rest/pkg/jsonutil"
	"github.com/stretchr/testify/assert"
)

func testHandle(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user.IsAnonymous() {
		j.Write(w, &j.Response{Code: http.StatusUnauthorized})
		return
	}
	if !user.IsAuthenticated() {
		j.Write(w, &j.Response{Code: http.StatusUnauthorized})
		return
	}
	j.Write(w, user)
}

func TestHandlerMiddleware(t *testing.T) {
	_, err := testHandler.db.ExecQuery(context.Background(), "DROP TABLE IF EXISTS auth_users")
	assert.Nil(t, err)
	_ = testHandler.setup()

	body := strings.NewReader(`{
			"username": "hello",
			"password": "world"
		}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	w := httptest.NewRecorder()
	testHandler.ServeHTTP(w, req)

	body = strings.NewReader(`{
		"username": "hello",
		"password": "world"
	}`)
	req = httptest.NewRequest(http.MethodPost, "/auth/login", body)
	w = httptest.NewRecorder()
	testHandler.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
	var resData map[string]string
	err = json.Unmarshal(data, &resData)
	assert.Nil(t, err)
	token := resData["token"]

	t.Run("not authorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Add(AuthorizationHeader, "Bearer "+token)
		w = httptest.NewRecorder()
		testHandle(w, req)
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})
	t.Run("authorized", func(t *testing.T) {
		middleware := NewMiddleware([]byte(testSecret))
		authHandler := middleware(http.HandlerFunc(testHandle))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Add(AuthorizationHeader, "Bearer "+token)
		authHandler.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)

		data, err = io.ReadAll(res.Body)
		if err != nil {
			t.Error(err)
		}
		t.Log("get user data middleware: ", string(data))
		var userRes map[string]any
		err = json.Unmarshal(data, &userRes)
		assert.Nil(t, err)
	})
}
