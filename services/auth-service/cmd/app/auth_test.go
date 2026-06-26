package main

import (
	"app/auth-service/internal/auth"
	"app/auth-service/internal/common"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"shared/shared_testing"
	"strconv"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

const (
	testPassword = "test_password123##@"
)

func TestRegisterSuccessful(t *testing.T) {
	dataSqlFile, errReadFile := os.ReadFile("load_mock_users.sql")
	if errReadFile != nil {
		t.Fatal("failed to read sql file: ", errReadFile)
	}
	shared_testing.RefreshUserTestData(dataSqlFile, t)
	deleteRedisData(t)
	deleteMailPitMessages(t)
	_, _, app := App()
	testServer := httptest.NewServer(app)
	defer testServer.Close()
	bodyRegisterUser := &auth.RequestRegister{
		Name:     "example_name",
		Email:    "exampleregister@gmail.com",
		Password: testPassword,
	}
	dataReq, errMarshal := json.Marshal(bodyRegisterUser)
	if errMarshal != nil {
		t.Fatal("failed to prepare request register: ", errMarshal)
	}
	respRegister, errResp := http.Post(testServer.URL+"/api/v1/register", "application/json", bytes.NewBuffer(dataReq))
	if errResp != nil {
		t.Fatal("failed to send request: ", errResp)
	}
	helperTestConfirmAndRefresh(respRegister, auth.ActionRegister, "", testServer, t)
}
func TestLoginSuccessful(t *testing.T) {
	dataSqlFile, errReadFile := os.ReadFile("load_mock_users.sql")
	if errReadFile != nil {
		t.Fatal("failed to read sql file: ", errReadFile)
	}
	shared_testing.RefreshUserTestData(dataSqlFile, t)
	deleteRedisData(t)
	deleteMailPitMessages(t)
	_, _, app := App()
	testServer := httptest.NewServer(app)
	defer testServer.Close()
	RequestRegisterUser := &auth.RequestLogin{
		Email:    "examplelogin@gmail.com",
		Password: testPassword,
	}
	dataReq, errMarshal := json.Marshal(RequestRegisterUser)
	if errMarshal != nil {
		t.Fatal("failed to prepare request login: ", errMarshal)
	}
	respLogin, errResp := http.Post(testServer.URL+"/api/v1/login", "application/json", bytes.NewBuffer(dataReq))
	if errResp != nil {
		t.Fatal("failed to get response: ", errResp)
	}
	helperTestConfirmAndRefresh(respLogin, auth.ActionLogin, "", testServer, t)
}

var CaseActionRecovery = []struct {
	auth.RequestRecovery
	Action      string
	NewPassword string
}{
	{Action: auth.ActionRecoveryPassword, NewPassword: "newPassword123", RequestRecovery: auth.RequestRecovery{Email: "examplerecoverypassword@gmail.com"}},
	{Action: auth.ActionRecoveryUser, RequestRecovery: auth.RequestRecovery{Email: "examplerecoveryuser@gmail.com", Password: testPassword}},
}

func TestRecoverySuccess(t *testing.T) {
	_, _, app := App()
	testServer := httptest.NewServer(app)
	for _, testCase := range CaseActionRecovery {
		dataSqlFile, errReadFile := os.ReadFile("load_mock_users.sql")
		if errReadFile != nil {
			t.Fatal("failed to read sql file: ", errReadFile)
		}
		shared_testing.RefreshUserTestData(dataSqlFile, t)
		deleteRedisData(t)
		deleteMailPitMessages(t)
		dataRecovery, errMarshalRecovery := json.Marshal(testCase.RequestRecovery)
		if errMarshalRecovery != nil {
			t.Fatal("failed to prepare request recovery: ", errMarshalRecovery)
		}
		respRecovery, errRespRecovery := http.Post(testServer.URL+"/api/v1/recovery?action="+testCase.Action, "application/json", bytes.NewBuffer(dataRecovery))
		if errRespRecovery != nil {
			t.Fatal("failed to get response: ", errRespRecovery)
		}
		helperTestConfirmAndRefresh(respRecovery, testCase.Action, testCase.NewPassword, testServer, t)
	}
}
func deleteMailPitMessages(t *testing.T) {
	req, errReq := http.NewRequest(http.MethodDelete, "http://localhost:8025/api/v1/messages", bytes.NewBuffer([]byte("{IDs:[]}")))
	if errReq != nil {
		t.Fatal("failed to delete Mailpit: ", errReq)
	}
	resp, errResp := http.DefaultClient.Do(req)
	if errResp != nil {
		t.Fatal("failed to get response Mailpit: ", errResp)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			t.Fatal("failed to close: ", errClose)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Mailpit expected %d got %d", http.StatusOK, resp.StatusCode)
	}
}
func deleteRedisData(t *testing.T) {
	if errFileEnvTest := godotenv.Load(".env.test"); errFileEnvTest != nil {
		t.Fatal(".env.test file not found")
	}
	redisAddress := os.Getenv("REDIS_ADDRESS")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisAddress == "" {
		t.Fatal("environment variable 'REDIS_ADDRESS' not found")
	}
	if redisPassword == "" {
		t.Fatal("environment variable 'REDIS_PASSWORD' not found")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: redisPassword,
	})
	defer func() {
		if errClose := rdb.Close(); errClose != nil {
			t.Fatal("failed to close rdb: ", errClose)
		}
	}()
	ctxTimeout, cancel := context.WithTimeout(context.Background(), common.CtxTimeout)
	defer cancel()
	if errFlushAll := rdb.FlushAll(ctxTimeout).Err(); errFlushAll != nil {
		t.Fatal("failed to flushAll rdb: ", errFlushAll)
	}
}
func helperTestConfirmAndRefresh(resp *http.Response, action, newPassword string, testServer *httptest.Server, t *testing.T) {
	resultDataLogin := shared_testing.HelperCheckResponse(resp, t)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected %d got %d", http.StatusAccepted, resp.StatusCode)
	}
	respReg, ok := resultDataLogin.(map[string]any)
	if !ok {
		t.Fatal("failed to assertion type (map[string)any): ", resultDataLogin)
	}
	dataConfirm := prepareConfirm(newPassword, t)
	reqConfirm, errReqConfirm := http.NewRequest(http.MethodPost, testServer.URL+"/api/v1/confirm?action="+action, bytes.NewBuffer(dataConfirm))
	if errReqConfirm != nil {
		t.Fatal("failed to prepare request confirm: ", errReqConfirm)
	}
	reqConfirm.Header.Set("X-Session-Token", "Bearer "+respReg["session_jwt"].(string))
	respConfirm, errRespConfirm := http.DefaultClient.Do(reqConfirm)
	if errRespConfirm != nil {
		t.Fatal("failed to get response confirm: ", errRespConfirm)
	}
	resultDataConfirm := shared_testing.HelperCheckResponse(respConfirm, t)
	var expectedStatusCode int
	if action == auth.ActionRegister {
		expectedStatusCode = http.StatusCreated
	} else {
		expectedStatusCode = http.StatusOK
	}
	if respConfirm.StatusCode != expectedStatusCode {
		t.Fatalf("expected register confirm %d got %d", expectedStatusCode, respConfirm.StatusCode)
	}
	dataRespConfirm, ok := resultDataConfirm.(map[string]any)
	if !ok {
		t.Fatal("failed to assertion type (map[string]any): ", dataRespConfirm)
	}
	bodyRefresh := auth.RequestRefresh{
		RefreshJwt: dataRespConfirm["refresh_jwt"].(string),
	}
	dataRefresh, errMarshalRefresh := json.Marshal(bodyRefresh)
	if errMarshalRefresh != nil {
		t.Fatal("failed to marshal refresh data: ", errMarshalRefresh)
	}
	respRefresh, errRespRefresh := http.Post(testServer.URL+"/api/v1/refresh", "application/json", bytes.NewBuffer(dataRefresh))
	if errRespRefresh != nil {
		t.Fatal("failed to get response refresh: ", errRespRefresh)
	}
	resultDataRefresh := shared_testing.HelperCheckResponse(respRefresh, t)
	if respRefresh.StatusCode != http.StatusOK {
		t.Fatalf("expected refresh %d got %d", http.StatusOK, respRefresh.StatusCode)
	}
	dataRespRefresh, ok := resultDataRefresh.(map[string]any)
	if !ok {
		t.Fatal("failed to assertion type (map[string]any): ", dataRespRefresh)
	}
	t.Log(dataRespRefresh["refresh_jwt"])
}
func prepareConfirm(password string, t *testing.T) []byte {
	respCode, errRespCode := http.Get("http://localhost:8025/api/v1/messages")
	if errRespCode != nil {
		t.Fatal("failed to send response code: ", errRespCode)
	}
	defer func() {
		if errClose := respCode.Body.Close(); errClose != nil {
			t.Fatal("failed to close: ", errClose)
		}
	}()
	dataCode, errRead := io.ReadAll(respCode.Body)
	if errRead != nil {
		t.Fatal("failed to read response data code; ", errRead)
	}
	reg, errRegexp := regexp.Compile(`code:\s\b(\d{6})\b`)
	if errRegexp != nil {
		t.Fatal("incorrect regexp: ", errRegexp)
	}
	regCode := reg.FindString(string(dataCode))
	codeStr := strings.Split(regCode, "code: ")
	if len(codeStr) != 2 {
		t.Fatal("failed to extract code: ", codeStr)
	}
	code, errParseCode := strconv.Atoi(codeStr[1])
	if errParseCode != nil {
		t.Fatal("failed to parse code: ", errParseCode)
	}
	requestConfirm := auth.RequestConfirm{
		Code:        code,
		NewPassword: password,
	}
	dataResp, errMarshal := json.Marshal(requestConfirm)
	if errMarshal != nil {
		t.Fatal("failed to prepare request confirm: ", errMarshal)
	}
	return dataResp
}
