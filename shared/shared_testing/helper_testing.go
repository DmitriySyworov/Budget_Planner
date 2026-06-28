package shared_testing

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func RefreshUserTestData(sqlFileData []byte, deleteTableNames []string, t *testing.T) {
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
	if _, errDelete := db.Exec("TRUNCATE " + strings.Join(deleteTableNames, ",")); errDelete != nil {
		t.Fatal("failed to delete old data: ", errDelete)
	}
	if _, errLoad := db.Exec(string(sqlFileData)); errLoad != nil {
		t.Fatal("failed to load new data: ", errLoad)
	}
}

type TestResponse[T any] struct {
	Success bool
	Data    T                 `json:"data,omitempty"`
	Error   map[string]string `json:"errors,omitempty"`
}

func HelperHandleResponse[T any](resp *http.Response, statusCode int, t *testing.T) T {
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			t.Fatal("failed to close response: ", errClose)
		}
	}()
	var payload TestResponse[T]
	if statusCode != http.StatusNoContent {
		if errDecode := json.NewDecoder(resp.Body).Decode(&payload); errDecode != nil {
			t.Fatal("failed to decode response: " + errDecode.Error())
		}
		if !payload.Success {
			t.Error("response isn't successful")
		}
	}
	if resp.StatusCode != statusCode {
		t.Fatalf("expected %d got %d", statusCode, resp.StatusCode)
	}
	return payload.Data
}
func CreateTestAccessToken(userUUID, signature string, t *testing.T) string {
	claim := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"type": "access",
		"sub":  userUUID,
		"exp":  time.Now().Add(time.Minute * 5).Unix(),
	})
	token, errToken := claim.SignedString([]byte(signature))
	if errToken != nil {
		t.Fatal("failed to sign token: ", errToken)
		return ""
	}
	return token
}
