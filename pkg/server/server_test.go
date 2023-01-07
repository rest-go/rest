package server

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		res, err := request(http.MethodGet, "/", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)

		res, err = request(http.MethodGet, "/customers", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)
	})

	t.Run("invalid table", func(t *testing.T) {
		res, err := request(http.MethodGet, "/0invalid_table_name", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, res.Code, res.Msg)
	})

	t.Run("invalid method", func(t *testing.T) {
		res, err := request(http.MethodHead, "/customers", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusMethodNotAllowed, res.Code, res.Msg)
	})
}

// Test Create and Delete together so that we can delete the data that just
// created without affect other tests to assert test data.
func TestServer_Create_Delete(t *testing.T) {
	t.Run("duplicate id", func(t *testing.T) {
		res, err := request(http.MethodPost, "/customers", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, res.Code, "empty body should return bad request")

		body := strings.NewReader(`{
			"CustomerId": 1,
			"FirstName": "first name",
			"LastName": "last_name",
			"Email": "a@b.com",
			"Active": true
		}`)
		res, err = request(http.MethodPost, "/customers", body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusConflict, res.Code, res.Msg)
	})
	t.Run("single", func(t *testing.T) {
		body := strings.NewReader(`{
			"CustomerId": 100,
			"FirstName": "first name",
			"LastName": "last_name",
			"Email": "a@b.com",
			"Active": true
		}`)
		res, err := request(http.MethodPost, "/customers", body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)
	})

	t.Run("bulk", func(t *testing.T) {
		body := strings.NewReader(`[
			{
				"InvoiceID": 100,
				"CustomerId":100,
				"InvoiceDate": "2023-01-02 03:04:05",
				"BillingAddress": "I'm an address",
				"Total":3.1415926,
				"Data": "{\"Country\": \"I'm an country\", \"PostalCode\":1234}"
			},
			{
				"InvoiceID": 101,
				"CustomerId":100,
				"InvoiceDate": "2023-01-02 03:04:05",
				"BillingAddress": "I'm an address",
				"Total":1.141421,
				"Data": "{\"Country\": \"I'm an country\", \"PostalCode\":1234}"
			}
		]`)
		res, err := request(http.MethodPost, "/invoices", body)
		assert.Nil(t, err)
		assert.Equal(t, 200, res.Code, res.Msg)
	})

	t.Run("delete", func(t *testing.T) {
		t.Log("delete customers created above")
		res, err := request(http.MethodDelete, "/customers?CustomerId=eq.100", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)

		res, err = request(http.MethodGet, "/customers?CustomerId=eq.100", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)
		objects := res.Data.([]any)
		assert.Equal(t, 0, len(objects))

		t.Log("delete invoices created above")
		res, err = request(http.MethodDelete, "/invoices?InvoiceID=in.(100,101)", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)

		res, err = request(http.MethodGet, "/invoices?InvoiceID=in.(100,101)", nil)
		assert.Nil(t, err)
		objects = res.Data.([]any)
		assert.Equal(t, 0, len(objects))
	})
}

func TestServer_Read(t *testing.T) {
	t.Run("one", func(t *testing.T) {
		res, err := request(http.MethodGet, "/customers?CustomerID=eq.1", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)
		objects, ok := res.Data.([]any)
		assert.True(t, ok)
		assert.Equal(t, 1, len(objects))
		t.Log("get customers: ", objects)
	})

	t.Run("one singular", func(t *testing.T) {
		res, err := request(http.MethodGet, "/invoices?InvoiceID=eq.1&singular", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)
		object, ok := res.Data.(map[string]any)
		assert.True(t, ok)
		t.Log("get invoice: ", object)
	})

	t.Run("many", func(t *testing.T) {
		res, err := request(http.MethodGet, "/invoices", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)
		objects, ok := res.Data.([]any)
		assert.True(t, ok)
		assert.Equal(t, 2, len(objects))
		t.Log("get invoices: ", objects)
	})

	t.Run("many with page", func(t *testing.T) {
		res, err := request(http.MethodGet, "/invoices?page=2&page_size=1", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)
		objects, ok := res.Data.([]any)
		assert.True(t, ok)
		assert.Equal(t, 1, len(objects))
		t.Log("get invoices: ", objects)
	})

	t.Run("many singular with error", func(t *testing.T) {
		res, err := request(http.MethodGet, "/invoices?singular", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, res.Code, res.Msg)
	})
}
func TestServerUpdate(t *testing.T) {
	newName := "I'm a new first name"
	body := strings.NewReader(fmt.Sprintf(`{
			"FirstName": %q
		}`, newName))
	res, err := request(http.MethodPut, "/customers?CustomerId=eq.1", body)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.Code, res.Msg)

	res, err = request(http.MethodGet, "/customers?CustomerId=eq.1&singular", body)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.Code, res.Msg)
	data := res.Data.(map[string]any)
	firstName := data["FirstName"].(string)
	assert.Equal(t, newName, firstName)
}

func TestServerDebug(t *testing.T) {
	res, err := request(http.MethodGet, "/customers?debug", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.Code, res.Msg)
	data := res.Data.(map[string]any)
	t.Log("get debug data: ", data["query"], data["args"])

	res, err = request(http.MethodDelete, "/customers?CustomerId=eq.1&debug", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.Code, res.Msg)
	data = res.Data.(map[string]any)
	t.Log("get debug data: ", data["query"], data["args"])
}
func TestServerCount(t *testing.T) {
	res, err := request(http.MethodGet, "/customers?count", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.Code, res.Msg)
	data := res.Data.(map[string]any)
	assert.Equal(t, float64(1), data["count"])

	res, err = request(http.MethodGet, "/invoices?count", nil)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.Code, res.Msg)
	data = res.Data.(map[string]any)
	assert.Equal(t, float64(2), data["count"])
}
