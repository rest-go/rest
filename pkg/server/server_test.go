package server

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rest-go/rest/pkg/auth"
)

func TestServer(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		code, _, err := request(http.MethodGet, "/", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)

		code, data, err := request(http.MethodGet, "/customers", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)
		assertLength(t, 1, data)
	})

	t.Run("invalid table", func(t *testing.T) {
		code, _, err := request(http.MethodGet, "/invalid_table_name", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusNotFound, code)
	})

	t.Run("invalid method", func(t *testing.T) {
		code, _, err := request(http.MethodHead, "/customers", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusMethodNotAllowed, code)
	})
}

// Test Create and Delete together so that we can delete the data that just
// created without affect other tests to assert test data.
func TestServer_Create_Delete(t *testing.T) {
	t.Run("duplicate id", func(t *testing.T) {
		code, _, err := request(http.MethodPost, "/customers", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, code)

		body := strings.NewReader(`{
			"Id": 1,
			"FirstName": "first name",
			"LastName": "last_name",
			"Email": "a@b.com",
			"Active": true
		}`)
		code, _, err = request(http.MethodPost, "/customers", body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusConflict, code)
	})
	t.Run("single", func(t *testing.T) {
		body := strings.NewReader(`{
			"Id": 100,
			"FirstName": "first name",
			"LastName": "last_name",
			"Email": "a@b.com",
			"Active": true
		}`)
		code, _, err := request(http.MethodPost, "/customers", body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)
	})

	t.Run("bulk with un-uniform keys", func(t *testing.T) {
		body := strings.NewReader(`[
			{
				"Id": 100,
				"CustomerId":100,
				"InvoiceDate": "2023-01-02 03:04:05",
				"BillingAddress": "I'm an address",
				"Total":3.1415926,
				"Data": "{\"Country\": \"I'm an country\", \"PostalCode\":1234}"
			},
			{
				"Id": 101,
				"InvoiceDate": "2023-01-02 03:04:05",
				"Data": "{\"Country\": \"I'm an country\", \"PostalCode\":1234}"
			}
		]`)
		code, _, err := request(http.MethodPost, "/invoices", body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, code)
	})

	t.Run("bulk", func(t *testing.T) {
		body := strings.NewReader(`[
			{
				"Id": 100,
				"CustomerId":100,
				"InvoiceDate": "2023-01-02 03:04:05",
				"BillingAddress": "I'm an address",
				"Total":3.1415926,
				"Data": "{\"Country\": \"I'm an country\", \"PostalCode\":1234}"
			},
			{
				"Id": 101,
				"CustomerId":100,
				"InvoiceDate": "2023-01-02 03:04:05",
				"BillingAddress": "I'm an address",
				"Total":1.141421,
				"Data": "{\"Country\": \"I'm an country\", \"PostalCode\":1234}"
			}
		]`)
		code, _, err := request(http.MethodPost, "/invoices", body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)
	})

	t.Run("delete without conditions is not allowed unless explicitly", func(t *testing.T) {
		code, _, err := request(http.MethodDelete, "/customers", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, code)
	})

	t.Run("delete", func(t *testing.T) {
		t.Log("delete customers created above")
		code, _, err := request(http.MethodDelete, "/customers/100", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)

		code, _, err = request(http.MethodGet, "/customers/100", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusNotFound, code)

		t.Log("delete invoices created above")
		code, _, err = request(http.MethodDelete, "/invoices?Id=in.(100,101)", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)

		code, data, err := request(http.MethodGet, "/invoices?Id=in.(100,101)", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)
		assertLength(t, 0, data)
	})
}

func TestServer_Read(t *testing.T) {
	t.Run("one", func(t *testing.T) {
		code, data, err := request(http.MethodGet, "/customers/1", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)
		assertEqualField(t, "1", data, "Id")
	})

	t.Run("no primary key", func(t *testing.T) {
		code, _, err := request(http.MethodGet, "/articles/1", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, code)
	})

	t.Run("invalid select", func(t *testing.T) {
		code, _, err := request(http.MethodGet, "/articles?select=title%3b", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, code)
	})

	t.Run("one singular", func(t *testing.T) {
		code, data, err := request(http.MethodGet, "/invoices/1?singular", nil)
		assert.Equal(t, http.StatusOK, code)
		assert.Nil(t, err)
		assertEqualField(t, "1", data, "Id")
	})

	t.Run("many", func(t *testing.T) {
		code, data, err := request(http.MethodGet, "/invoices", nil)
		assert.Equal(t, http.StatusOK, code)
		assert.Nil(t, err)
		assertLength(t, 2, data)
		t.Log("get invoices: ", data)
	})

	t.Run("many with order", func(t *testing.T) {
		code, data, err := request(http.MethodGet, "/invoices?order=Id.desc", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)
		assertLength(t, 2, data)
		t.Log("get invoices: ", data)
	})

	t.Run("many with page", func(t *testing.T) {
		code, data, err := request(http.MethodGet, "/invoices", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)
		assertLength(t, 2, data)

		code, data, err = request(http.MethodGet, "/invoices?page=2&page_size=1", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)
		assertLength(t, 1, data)
	})

	t.Run("many singular with error", func(t *testing.T) {
		code, _, err := request(http.MethodGet, "/invoices?singular", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, code)
	})
}
func TestServerUpdate(t *testing.T) {
	newName := "I'm a new first name"
	body := strings.NewReader(fmt.Sprintf(`{
			"FirstName": %q
		}`, newName))
	code, _, err := request(http.MethodPut, "/customers/1", body)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)

	code, data, err := request(http.MethodGet, "/customers/1", body)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	assertEqualField(t, newName, data, "FirstName")

	t.Run("update without conditions is not allowed unless explicitly", func(t *testing.T) {
		newName := "update without condition"
		body := strings.NewReader(fmt.Sprintf(`{
			"FirstName": %q
		}`, newName))
		code, _, err := request(http.MethodPut, "/customers", body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, code)

		body = strings.NewReader(fmt.Sprintf(`{
			"FirstName": %q
		}`, newName))
		code, _, err = request(http.MethodPut, "/customers?1=eq.1", body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)
	})
}

func TestServerDebug(t *testing.T) {
	code, data, err := request(http.MethodGet, "/customers?debug", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	m := data.(map[string]any)
	t.Log("get debug data: ", m["query"], m["args"])

	code, data, err = request(http.MethodDelete, "/customers/1&debug", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	m = data.(map[string]any)
	t.Log("get debug data: ", m["query"], m["args"])
}
func TestServerCount(t *testing.T) {
	code, data, err := request(http.MethodGet, "/customers?count", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	t.Log("get data: ", data)
	count := data.(float64)
	assert.Equal(t, float64(1), count, data)

	code, data, err = request(http.MethodGet, "/invoices?count", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	count = data.(float64)
	assert.Equal(t, float64(2), count, data)

	code, data, err = request(http.MethodGet, "/invoices?id=eq.1&count", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	count = data.(float64)
	assert.Equal(t, float64(1), count, data)
}

func TestServerAuth(t *testing.T) {
	s := New(&DBConfig{URL: "sqlite://ci.db"}, EnableAuth(true))
	defer s.Close()

	middleware := auth.NewMiddleware([]byte("test-secret"))
	authServer := middleware(s)

	t.Run("anonymous user doesn't have permission on articles", func(t *testing.T) {
		code, _, err := requestHandler(authServer, "", http.MethodGet, "/articles", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusUnauthorized, code)

		code, _, err = requestHandler(authServer, "", http.MethodGet, "/articles?mine", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusUnauthorized, code)
	})

	t.Run("authenticated user has permission on all methods", func(t *testing.T) {
		token, err := auth.GenJWTToken([]byte("test-secret"), map[string]any{"user_id": 1})
		if err != nil {
			t.Error(err)
		}
		code, _, err := requestHandler(authServer, token, http.MethodGet, "/articles", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)

		body := strings.NewReader(`{
		"title": "I'm a title"
	}`)
		code, _, err = requestHandler(authServer, token, http.MethodPost, "/articles", body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)

		body = strings.NewReader(`{
		"title": "I'm another title"
	}`)
		code, _, err = requestHandler(authServer, token, http.MethodPut, "/articles?title=eq.title", body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)

		code, _, err = requestHandler(authServer, token, http.MethodDelete, "/articles?title=eq.title", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)
	})

	t.Run("only admin users have permission on auth_policies", func(t *testing.T) {
		token, err := auth.GenJWTToken([]byte("test-secret"), map[string]any{"user_id": 1})
		if err != nil {
			t.Error(err)
		}
		code, _, err := requestHandler(authServer, token, http.MethodGet, "/auth_policies", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusForbidden, code)

		token, err = auth.GenJWTToken([]byte("test-secret"), map[string]any{"user_id": 1, "is_admin": true})
		if err != nil {
			t.Error(err)
		}
		code, _, err = requestHandler(authServer, token, http.MethodGet, "/articles?mine", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, code)
	})
}
