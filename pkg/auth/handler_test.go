package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
		w := httptest.NewRecorder()
		testHandler.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
	})

	t.Run("action not provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/", nil)
		w := httptest.NewRecorder()
		testHandler.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("action not supported", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/x", nil)
		w := httptest.NewRecorder()
		testHandler.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func TestHandlerActions(t *testing.T) {
	t.Run("register", func(t *testing.T) {
		body := strings.NewReader(`{
			"username": "hello",
			"password": "world"
		}`)
		req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
		w := httptest.NewRecorder()
		testHandler.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)

		body = strings.NewReader(`{
			"username": "hello",
			"password": "world"
		}`)
		t.Log("register same username twice, should return error")
		req = httptest.NewRequest(http.MethodPost, "/auth/register", body)
		w = httptest.NewRecorder()
		testHandler.ServeHTTP(w, req)
		res = w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusConflict, res.StatusCode)
	})

	t.Run("login", func(t *testing.T) {
		body := strings.NewReader(`{
			"username": "hello",
			"password": "world"
		}`)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
		w := httptest.NewRecorder()
		testHandler.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
		data, err := io.ReadAll(res.Body)
		if err != nil {
			t.Error(err)
		}

		var resData map[string]string
		err = json.Unmarshal(data, &resData)
		assert.Nil(t, err)
		t.Log("get token: ", resData["token"])

		t.Log("login with wrong password")
		body = strings.NewReader(`{
			"username": "hello",
			"password": "world2"
		}`)
		req = httptest.NewRequest(http.MethodPost, "/auth/login", body)
		w = httptest.NewRecorder()
		testHandler.ServeHTTP(w, req)
		res = w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

		t.Log("login with wrong username")
		body = strings.NewReader(`{
			"username": "hello2",
			"password": "world"
		}`)
		req = httptest.NewRequest(http.MethodPost, "/auth/login", body)
		w = httptest.NewRecorder()
		testHandler.ServeHTTP(w, req)
		res = w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("logout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		w := httptest.NewRecorder()
		testHandler.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}
