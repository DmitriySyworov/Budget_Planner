package shared_testing

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"shared/response"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func RefreshUserTestData(sqlFileData []byte, t *testing.T) {
	if errFileEnvTest := godotenv.Load(".env.test"); errFileEnvTest != nil {
		t.Fatal(".env.test file not found")
	}
	dsnTest := os.Getenv("DSN")
	if dsnTest == "" {
		t.Fatal("environment variable 'DSN' not found")
	}
	db, errOpen := sql.Open("postgres", dsnTest)
	defer func() {
		if errClose := db.Close(); errClose != nil {
			t.Fatal("failed to close sql driver")
		}
	}()
	if errOpen != nil {
		t.Fatal("failed to connect PostgreSQL: ", errOpen)
	}
	if _, errDelete := db.Exec("TRUNCATE users"); errDelete != nil {
		t.Fatal("failed to delete old data: ", errDelete)
	}
	if _, errLoad := db.Exec(string(sqlFileData)); errLoad != nil {
		t.Fatal("failed to load new data: ", errLoad)
	}
}
func HelperCheckResponse(resp *http.Response, t *testing.T) any {
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			t.Fatal("failed to close response: ", errClose)
		}
	}()
	var payload response.Response
	if errDecode := json.NewDecoder(resp.Body).Decode(&payload); errDecode != nil {
		t.Fatal("failed to decode response: " + errDecode.Error())
	}
	if !payload.Success {
		t.Fatal("response isn't successful")
	}
	return payload.Data
}
