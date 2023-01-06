package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/shellfly/rest/pkg/database"
)

var testServer *Server

var setupSQL = `
DROP TABLE IF EXISTS "customers";
CREATE TABLE IF NOT EXISTS "customers"
(
    [CustomerId] INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    [FirstName] NVARCHAR(40)  NOT NULL,
    [LastName] NVARCHAR(20)  NOT NULL,
    [Email] NVARCHAR(60)  NOT NULL,
	[Active] BOOL NOT NULL
);
DROP TABLE IF EXISTS "invoices";
CREATE TABLE IF NOT EXISTS "invoices"
(
    [InvoiceId] INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    [CustomerId] INTEGER  NOT NULL,
    [InvoiceDate] DATETIME  NOT NULL,
    [BillingAddress] NVARCHAR(70),
    [Total] NUMERIC(10,2)  NOT NULL,
	[Data] JSON NOT NULL,
    FOREIGN KEY ([CustomerId]) REFERENCES "customers" ([CustomerId])
                ON DELETE NO ACTION ON UPDATE NO ACTION
);
CREATE INDEX [IFK_InvoiceCustomerId] ON "invoices" ([CustomerId]);
`

func TestMain(m *testing.M) {
	testServer = NewServer("sqlite://ci.db")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if _, err := database.ExecQuery(ctx, testServer.db, setupSQL); err != nil {
		log.Fatal("setup database error", err)
	}
	code := m.Run()
	cancel()
	os.Exit(code)
}

func request(method, target string, body io.Reader) (*Response, error) {
	req := httptest.NewRequest(method, target, body)
	w := httptest.NewRecorder()
	testServer.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response Response
	err = json.Unmarshal(data, &response)
	return &response, err
}

func TestServer(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		res, err := request(http.MethodPost, "/customers", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, res.Code, "empty body should return bad request")

		body := strings.NewReader(`{
			"CustomerId": 1,
			"FirstName": "first name",
			"LastName": "last_name",
			"Email": "a@b.com", 
			"Active":true
		}`)
		res, err = request(http.MethodPost, "/customers", body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)

		body = strings.NewReader(`{
			"CustomerId": 1,
			"FirstName": "first name",
			"LastName": "last_name",
			"Email": "a@b.com",
			"Active": true
		}`)
		res, err = request(http.MethodPost, "/customers", body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusConflict, res.Code, "duplicate customer id should return conflict code")

		body = strings.NewReader(`[
			{
				"InvoiceID": 1,
				"CustomerId":1,
				"InvoiceDate": "2023-01-02 03:04:05",
				"BillingAddress": "I'm an address",
				"Total":3.1415926,
				"Data": "{\"Country\": \"I'm an country\", \"PostalCode\":1234}"
			},
			{
				"InvoiceID": 2,
				"CustomerId":1,
				"InvoiceDate": "2023-01-02 03:04:05",
				"BillingAddress": "I'm an address",
				"Total":1.141421,
				"Data": "{\"Country\": \"I'm an country\", \"PostalCode\":1234}"
			}
		]`)
		res, err = request(http.MethodPost, "/invoices", body)
		assert.Nil(t, err)
		assert.Equal(t, 200, res.Code, res.Msg)
	})

	t.Run("Read", func(t *testing.T) {
		res, err := request(http.MethodGet, "/customers", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)
		objects, ok := res.Data.([]any)
		assert.True(t, ok)
		assert.Equal(t, 1, len(objects))
		t.Log("get customers: ", objects)

		res, err = request(http.MethodGet, "/invoices", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)
		objects, ok = res.Data.([]any)
		assert.True(t, ok)
		assert.Equal(t, 2, len(objects))
		t.Log("get invoices: ", objects)
	})

	t.Run("Update", func(t *testing.T) {
		body := strings.NewReader(`{
			"FirstName": "I'm a new first name"
		}`)
		res, err := request(http.MethodPut, "/customers?CustomerId=eq.1", body)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)
	})

	t.Run("Delete", func(t *testing.T) {
		res, err := request(http.MethodDelete, "/customers?CustomerId=eq.1", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)

		res, err = request(http.MethodGet, "/customers", nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.Code, res.Msg)
		objects, ok := res.Data.([]any)
		assert.True(t, ok)
		assert.Equal(t, 0, len(objects))
	})
}
