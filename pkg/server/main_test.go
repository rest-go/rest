package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rest-go/rest/pkg/sqlx"
)

const (
	setupSQL = `
DROP TABLE IF EXISTS "customers";
CREATE TABLE IF NOT EXISTS "customers"
(
    [Id] INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    [FirstName] NVARCHAR(40)  NOT NULL,
    [LastName] NVARCHAR(20)  NOT NULL,
    [Email] NVARCHAR(60)  NOT NULL,
	[Active] BOOL NOT NULL
);
DROP TABLE IF EXISTS "invoices";
CREATE TABLE IF NOT EXISTS "invoices"
(
    [Id] INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
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
)

var testServer *Server

func TestMain(m *testing.M) {
	testServer = NewServer("sqlite://ci.db")
	if err := setupData(); err != nil {
		log.Fatal("setupData error: ", err)
	}
	os.Exit(m.Run())
}

func setupData() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := sqlx.ExecQuery(ctx, testServer.db, setupSQL); err != nil {
		return err
	}
	body := strings.NewReader(`{
		"Id": 1,
		"FirstName": "first name",
		"LastName": "last_name",
		"Email": "a@b.com", 
		"Active":true
	}`)
	code, data, err := request(http.MethodPost, "/customers", body)
	if err != nil || code != http.StatusOK {
		return fmt.Errorf("failed to insert customers, err: %w, code: %d, data: %v", err, code, data)
	}

	body = strings.NewReader(`[
			{
				"Id": 1,
				"CustomerId":1,
				"InvoiceDate": "2023-01-02 03:04:05",
				"BillingAddress": "I'm an address",
				"Total":3.1415926,
				"Data": "{\"Country\": \"I'm an country\", \"PostalCode\":1234}"
			},
			{
				"Id": 2,
				"CustomerId":1,
				"InvoiceDate": "2023-01-02 03:04:05",
				"BillingAddress": "I'm an address",
				"Total":1.141421,
				"Data": "{\"Country\": \"I'm an country\", \"PostalCode\":1234}"
			}
		]`)
	code, data, err = request(http.MethodPost, "/invoices", body)
	if err != nil || code != http.StatusOK {
		return fmt.Errorf("failed to insert invoices, err: %w, code: %d, data: %v", err, code, data)
	}
	return nil
}

func request(method, target string, body io.Reader) (code int, resData any, err error) {
	req := httptest.NewRequest(method, target, body)
	w := httptest.NewRecorder()
	testServer.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, nil, err
	}

	err = json.Unmarshal(data, &resData)
	return res.StatusCode, resData, err
}

func assertLength(t *testing.T, length int, data any) {
	t.Helper()
	objects := data.([]any)
	if length != len(objects) {
		t.Errorf("expected length: %d, got: %d", length, len(objects))
	}
}

func assertEqualField(t *testing.T, expectedVal string, data any, field string) {
	t.Helper()
	m := data.(map[string]any)
	if expectedVal != fmt.Sprintf("%v", m[field]) {
		t.Errorf("expected field val: %s, got: %v, data: %v", expectedVal, m[field], data)
	}
}
