package main

import (
	"app/auth-service/internal/common"
	"app/auth-service/internal/model"
	"app/auth-service/internal/user"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"shared/shared_common"
	"shared/shared_testing"
	"testing"
)

func TestGetUserSuccessful(t *testing.T) {
	const userGetUUID = "7b3e1f4a-6d2c-4b8a-9e1c-5f6a7b8c9d0e"
	dataSqlFile, errReadFile := os.ReadFile("load_mock_users.sql")
	if errReadFile != nil {
		t.Fatal("failed to read sql file: ", errReadFile)
	}
	shared_testing.RefreshUserTestData(dataSqlFile, []string{"users"}, t)
	confApi, _, app := App()
	accessToken := shared_testing.CreateTestAccessToken(userGetUUID, confApi.Signature, t)
	testServer := httptest.NewServer(app)
	defer testServer.Close()
	request, errReq := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/user", nil)
	if errReq != nil {
		t.Fatal("failed to prepare request: ", errReq)
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	respGet, errRespGet := http.DefaultClient.Do(request)
	if errRespGet != nil {
		t.Fatal("failed to get response: ", errRespGet)
	}
	userResp := shared_testing.HelperHandleResponse[user.ResponseUser](respGet, http.StatusOK, t)
	if userResp.UserUUID != userGetUUID {
		t.Fatalf("expected uuid %s got %s", userGetUUID, userResp.UserUUID)
	}
}

const (
	NewName     = "newName"
	NewEmail    = "newemail@gmail.com"
	NewPassword = "newpassword234Qw2"
	Email       = "exampleupdate@gmail.com"
)

var CaseDataUpdate = []user.RequestUpdateUser{
	{NewName: NewName},
	{NewName: NewName, NewPassword: NewPassword, Password: testPassword, Email: Email},
	{NewPassword: NewPassword, Password: testPassword, Email: Email},
	{NewEmail: NewEmail, NewPassword: NewPassword, Password: testPassword},
	{NewName: NewName, NewEmail: NewEmail, NewPassword: NewPassword, Password: testPassword},
}

func TestUpdateUserSuccessful(t *testing.T) {
	const userUpdateUUID = "f7b3a4c1-8d2e-4b9a-9e1c-5f6a7b8c9d0e"
	confApi, _, app := App()
	testServer := httptest.NewServer(app)
	defer testServer.Close()
	dataSqlFile, errReadFile := os.ReadFile("load_mock_users.sql")
	if errReadFile != nil {
		t.Fatal("failed to read sql file: ", errReadFile)
	}
	for _, test := range CaseDataUpdate {
		shared_testing.RefreshUserTestData(dataSqlFile, []string{"users"}, t)
		deleteRedisData(t)
		deleteMailPitMessages(t)
		accessToken := shared_testing.CreateTestAccessToken(userUpdateUUID, confApi.Signature, t)
		data, errMarshalUpdate := json.Marshal(test)
		if errMarshalUpdate != nil {
			t.Fatal("failed to prepare request: ", errMarshalUpdate)
		}
		request, errReq := http.NewRequest(http.MethodPatch, testServer.URL+"/api/v1/user", bytes.NewBuffer(data))
		if errReq != nil {
			t.Fatal("failed to prepare request: ", errReq)
		}
		request.Header.Set("Authorization", "Bearer "+accessToken)
		respUpdate, errRespUpdate := http.DefaultClient.Do(request)
		if errRespUpdate != nil {
			t.Fatal("failed to get response: ", errRespUpdate)
		}
		if test.NewName != "" && test.NewEmail == "" && test.NewPassword == "" {
			dataResp := shared_testing.HelperHandleResponse[model.Users](respUpdate, http.StatusOK, t)
			if dataResp.Name != NewName {
				t.Fatalf("expected name %s got %s", NewName, dataResp.Name)
			}
		} else {
			dataResp := shared_testing.HelperHandleResponse[common.ResponseAuth](respUpdate, http.StatusAccepted, t)
			if dataResp.SessionJwt == "" {
				t.Fatal("sessionJwt is empty")
			}
			code := helperExtractCode(t)
			bodyConfirm := user.RequestConfirm{
				Code: code,
			}
			dataConfirm, errMarshalConfirm := json.Marshal(bodyConfirm)
			if errMarshalConfirm != nil {
				t.Fatal("failed to prepare request Confirm: ", errMarshalConfirm)
			}
			requestConfirm, errReqConfirm := http.NewRequest(http.MethodPost, testServer.URL+"/api/v1/user/confirm?action=update", bytes.NewBuffer(dataConfirm))
			if errReqConfirm != nil {
				t.Fatal("failed to prepare request: ", errReqConfirm)
			}
			requestConfirm.Header.Set("X-Session-Token", "Bearer "+dataResp.SessionJwt)
			requestConfirm.Header.Set("Authorization", "Bearer "+accessToken)
			respConfirm, errRespConfirm := http.DefaultClient.Do(requestConfirm)
			if errRespConfirm != nil {
				t.Fatal("failed to get response confirm: ", errRespConfirm)
			}
			userUpdate := shared_testing.HelperHandleResponse[user.ResponseUser](respConfirm, http.StatusOK, t)
			if userUpdate.UserUUID != userUpdateUUID {
				t.Fatalf("expected user_uuid %s got %s", userUpdateUUID, userUpdate.UserUUID)
			}
		}
	}
}

var CaseDataRemove = []struct {
	Type     string
	UserUUID string
	user.RequestRemoveUser
}{
	{RequestRemoveUser: user.RequestRemoveUser{Email: "exampledelete@gmail.com", Password: testPassword}, Type: shared_common.TypeHardDelete, UserUUID: "5a1f4b3e-2c7d-491c-a3f5-6b2d8e1c9a4f"},
	{RequestRemoveUser: user.RequestRemoveUser{Email: "exampleremove@gmail.com", Password: testPassword}, Type: shared_common.TypeSoftDelete, UserUUID: "9f8e7d6c-5b4a-4321-a1b2-c3d4e5f6a7b8"},
}

func TestRemoveUserSuccessful(t *testing.T) {
	confApi, _, app := App()
	testServer := httptest.NewServer(app)
	defer testServer.Close()
	dataSqlFile, errReadFile := os.ReadFile("load_mock_users.sql")
	if errReadFile != nil {
		t.Fatal("failed to read sql file: ", errReadFile)
	}
	for _, test := range CaseDataRemove {
		shared_testing.RefreshUserTestData(dataSqlFile, []string{"users"}, t)
		deleteRedisData(t)
		deleteMailPitMessages(t)
		accessToken := shared_testing.CreateTestAccessToken(test.UserUUID, confApi.Signature, t)
		data, errMarshalRemove := json.Marshal(test.RequestRemoveUser)
		if errMarshalRemove != nil {
			t.Fatal("failed to prepare request: ", errMarshalRemove)
		}
		request, errReq := http.NewRequest(http.MethodDelete, testServer.URL+"/api/v1/user?type="+test.Type, bytes.NewBuffer(data))
		if errReq != nil {
			t.Fatal("failed to prepare request: ", errReq)
		}
		request.Header.Set("Authorization", "Bearer "+accessToken)
		respDelete, errRespDelete := http.DefaultClient.Do(request)
		if errRespDelete != nil {
			t.Fatal("failed to get response: ", errRespDelete)
		}
		dataResp := shared_testing.HelperHandleResponse[common.ResponseAuth](respDelete, http.StatusAccepted, t)
		if dataResp.SessionJwt == "" {
			t.Fatal("sessionJwt is empty")
		}
		code := helperExtractCode(t)
		bodyConfirm := user.RequestConfirm{
			Code: code,
		}
		dataConfirm, errMarshalConfirm := json.Marshal(bodyConfirm)
		if errMarshalConfirm != nil {
			t.Fatal("failed to prepare request Confirm: ", errMarshalConfirm)
		}
		requestConfirm, errReqConfirm := http.NewRequest(http.MethodPost, testServer.URL+"/api/v1/user/confirm?action="+test.Type, bytes.NewBuffer(dataConfirm))
		if errReqConfirm != nil {
			t.Fatal("failed to prepare request: ", errReqConfirm)
		}
		requestConfirm.Header.Set("X-Session-Token", "Bearer "+dataResp.SessionJwt)
		requestConfirm.Header.Set("Authorization", "Bearer "+accessToken)
		respConfirm, errRespConfirm := http.DefaultClient.Do(requestConfirm)
		if errRespConfirm != nil {
			t.Fatal("failed to get response confirm: ", errRespConfirm)
		}
		shared_testing.HelperHandleResponse[struct{}](respConfirm, http.StatusNoContent, t)
	}
}
