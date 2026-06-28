package main

import (
	"app/budget-planner/internal/budget"
	"app/budget-planner/internal/model"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"shared/shared_testing"
	"testing"

	"github.com/google/uuid"
)

func TestCreateBudgetSuccessful(t *testing.T) {
	conf, _, router := App()
	accessToken := shared_testing.CreateTestAccessToken("2e4b3c1d-8f9a-4c2b-b5e1-d3a7f8c9e0b2", conf.Signature, t)
	testServer := httptest.NewServer(router)
	defer testServer.Close()
	dataQuery, errReadFileSql := os.ReadFile("load-mock-budget-data.sql")
	if errReadFileSql != nil {
		t.Fatal("failed to read file sql: ", errReadFileSql)
	}
	shared_testing.RefreshUserTestData(dataQuery, []string{"budgets", "expenses", "description_expenses"}, t)
	bodyCreateBudget := budget.RequestCreateBudget{
		Amount:      "1234.54",
		Start:       "2026-06-01",
		Finish:      "2026-07-02",
		Description: "It's a very important budget plan",
	}
	dataCreate, errMarshalCreate := json.Marshal(bodyCreateBudget)
	if errMarshalCreate != nil {
		t.Fatal("failed to prepare body: ", errMarshalCreate)
	}
	requestCreate, errReqCreate := http.NewRequest(http.MethodPost, testServer.URL+"/api/v1/budget", bytes.NewBuffer(dataCreate))
	if errReqCreate != nil {
		t.Fatal("failed to prepare request: ", errReqCreate)
	}
	requestCreate.Header.Set("Authorization", "Bearer "+accessToken)
	respCreate, errRespCreate := http.DefaultClient.Do(requestCreate)
	if errRespCreate != nil {
		t.Fatal("failed to get response: ", errRespCreate)
	}
	dataRespCreate := shared_testing.HelperHandleResponse[model.Budgets](respCreate, http.StatusCreated, t)
	if _, errUUID := uuid.Parse(dataRespCreate.BudgetUUID); errUUID != nil {
		t.Fatal("incorrect budget_uuid: ", errUUID)
	}
}
func TestGetBudgetSuccessful(t *testing.T) {
	conf, _, router := App()
	accessToken := shared_testing.CreateTestAccessToken("6e5f4a3b-2c1d-4e9f-8a7b-6c5d4e3f2a1b", conf.Signature, t)
	testServer := httptest.NewServer(router)
	defer testServer.Close()
	dataQuery, errReadFileSql := os.ReadFile("load-mock-budget-data.sql")
	if errReadFileSql != nil {
		t.Fatal("failed to read file sql: ", errReadFileSql)
	}
	budgetUUID := "0f1e2d3c-4b5a-4678-9abc-def012345678"
	shared_testing.RefreshUserTestData(dataQuery, []string{"budgets", "expenses", "description_expenses"}, t)
	requestGet, errReqGet := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/budget/"+budgetUUID, nil)
	if errReqGet != nil {
		t.Fatal("failed to prepare request: ", errReqGet)
	}
	requestGet.Header.Set("Authorization", "Bearer "+accessToken)
	respGet, errRespGet := http.DefaultClient.Do(requestGet)
	if errRespGet != nil {
		t.Fatal("failed to get response: ", errRespGet)
	}
	dataRespGet := shared_testing.HelperHandleResponse[model.Budgets](respGet, http.StatusOK, t)
	if _, errUUID := uuid.Parse(dataRespGet.BudgetUUID); errUUID != nil {
		t.Fatal("incorrect budget_uuid: ", errUUID)
	}
}

func TestListBudgetSuccessful(t *testing.T) {
	conf, _, router := App()
	const userListUUID = "c9b8a7d6-e5f4-4321-890a-bcdef1234567"
	accessToken := shared_testing.CreateTestAccessToken(userListUUID, conf.Signature, t)
	testServer := httptest.NewServer(router)
	defer testServer.Close()
	dataQuery, errReadFileSql := os.ReadFile("load-mock-budget-data.sql")
	if errReadFileSql != nil {
		t.Fatal("failed to read file sql: ", errReadFileSql)
	}
	shared_testing.RefreshUserTestData(dataQuery, []string{"budgets", "expenses", "description_expenses"}, t)
	requestGet, errReqGet := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/budget", nil)
	if errReqGet != nil {
		t.Fatal("failed to prepare request: ", errReqGet)
	}
	requestGet.Header.Set("Authorization", "Bearer "+accessToken)
	respGet, errRespGet := http.DefaultClient.Do(requestGet)
	if errRespGet != nil {
		t.Fatal("failed to get response: ", errRespGet)
	}
	dataRespGet := shared_testing.HelperHandleResponse[[]model.Budgets](respGet, http.StatusOK, t)
	if len(dataRespGet) != 2 {
		t.Fatalf("expected len list budget %d got %d", 2, len(dataRespGet))
	}
	userFirstUUID := dataRespGet[0].UserUUID
	userSecondUUID := dataRespGet[1].UserUUID
	if _, errUUID := uuid.Parse(userFirstUUID); errUUID != nil {
		t.Error("incorrect uuid: ", errUUID)
	}
	if _, errUUID := uuid.Parse(userSecondUUID); errUUID != nil {
		t.Error("incorrect uuid: ", errUUID)
	}
	if _, errUUID := uuid.Parse(dataRespGet[0].BudgetUUID); errUUID != nil {
		t.Error("incorrect uuid: ", errUUID)
	}
	if _, errUUID := uuid.Parse(dataRespGet[1].BudgetUUID); errUUID != nil {
		t.Error("incorrect uuid: ", errUUID)
	}
	if userFirstUUID != userListUUID || userSecondUUID != userListUUID {
		t.Fatal("user_uuid does not match in records")
	}
}
