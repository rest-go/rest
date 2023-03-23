package auth

import (
	"context"

	"net/http"
	"os"
	"testing"

	j "github.com/rest-go/rest/pkg/jsonutil"
	"github.com/rest-go/rest/pkg/log"
)

var testHandler *Handler

const testSecret = "test-secret"

func TestMain(m *testing.M) {
	var err error
	testHandler, err = NewHandler("sqlite://ci.db", []byte(testSecret))
	if err != nil {
		log.Fatal(err)
	}

	// drop previous test tables
	_, err = testHandler.db.ExecQuery(context.Background(), "DROP TABLE IF EXISTS auth_users")
	if err != nil {
		log.Fatal(err)
	}
	_, err = testHandler.db.ExecQuery(context.Background(), "DROP TABLE IF EXISTS auth_policies")
	if err != nil {
		log.Fatal(err)
	}

	// setup auth tables
	val := testHandler.setup()
	if res, ok := val.(*j.Response); ok {
		if res.Code != http.StatusOK {
			log.Fatal(res.Msg)
		}
	}

	os.Exit(m.Run())
}
