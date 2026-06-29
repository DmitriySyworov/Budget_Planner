package main

import (
	"app/budget-planner/internal/expense"
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

func TestCreateExpenseSuccessful(t *testing.T) {
	const (
		budgetCreateUUID = "1a2b3c4d-5e6f-47a8-b9c0-1d2e3f4a5b6c"
		userCreateUUID   = "4a5b6c7d-8e9f-40a1-b2c3-d4e5f6a7b8c9"
	)
	conf, _, router := App()
	accessToken := shared_testing.CreateTestAccessToken(userCreateUUID, conf.Signature, t)
	testServer := httptest.NewServer(router)
	defer testServer.Close()
	dataQuery, errReadFileSql := os.ReadFile("load-mock-budget-data.sql")
	if errReadFileSql != nil {
		t.Fatal("failed to read file sql: ", errReadFileSql)
	}
	shared_testing.RefreshUserTestData(dataQuery, []string{"budgets", "expenses", "description_expenses"}, t)
	bodyCreateExpense := expense.CreateAndUpdateExpense{
		Category:    "health",
		Expense:     "234.78",
		Description: "buy pills",
	}
	dataCreate, errMarshalCreate := json.Marshal(bodyCreateExpense)
	if errMarshalCreate != nil {
		t.Fatal("failed to prepare body: ", errMarshalCreate)
	}
	requestCreate, errReqCreate := http.NewRequest(http.MethodPost, testServer.URL+"/api/v1/expense/"+budgetCreateUUID, bytes.NewBuffer(dataCreate))
	if errReqCreate != nil {
		t.Fatal("failed to prepare request: ", errReqCreate)
	}
	requestCreate.Header.Set("Authorization", "Bearer "+accessToken)
	respCreate, errRespCreate := http.DefaultClient.Do(requestCreate)
	if errRespCreate != nil {
		t.Fatal("failed to get response: ", errRespCreate)
	}
	dataRespCreate := shared_testing.HelperHandleResponse[model.DescriptionExpenses](respCreate, http.StatusCreated, t)
	if _, errUUID := uuid.Parse(dataRespCreate.Expense); errUUID != nil {
		t.Fatal("incorrect expense_uuid: ", errUUID)
	}
}
func TestGetDescriptionExpenseSuccessful(t *testing.T) {
	const (
		budgetUUID             = "1a2b3c4d-5e6f-47a8-b9c0-1d2e3f4a5b6c"
		expenseUUID            = "9926d83a-4be4-4298-ba98-25081b29cc36"
		descriptionExpenseUUID = "5b8b9333-d922-4a00-bf86-53d368e734bc"
		userUUID               = "4a5b6c7d-8e9f-40a1-b2c3-d4e5f6a7b8c9"
	)
	conf, _, router := App()
	accessToken := shared_testing.CreateTestAccessToken(userUUID, conf.Signature, t)
	testServer := httptest.NewServer(router)
	defer testServer.Close()
	dataQuery, errReadFileSql := os.ReadFile("load-mock-budget-data.sql")
	if errReadFileSql != nil {
		t.Fatal("failed to read file sql: ", errReadFileSql)
	}
	shared_testing.RefreshUserTestData(dataQuery, []string{"budgets", "expenses", "description_expenses"}, t)
	requestGet, errReqGet := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/expense/"+budgetUUID+"/"+descriptionExpenseUUID, nil)
	if errReqGet != nil {
		t.Fatal("failed to prepare request: ", errReqGet)
	}
	requestGet.Header.Set("Authorization", "Bearer "+accessToken)
	respRemove, errRespGet := http.DefaultClient.Do(requestGet)
	if errRespGet != nil {
		t.Fatal("failed to get response: ", errRespGet)
	}
	dataRespGet := shared_testing.HelperHandleResponse[model.DescriptionExpenses](respRemove, http.StatusOK, t)
	if _, errUUID := uuid.Parse(dataRespGet.DescriptionExpenseUUID); errUUID != nil {
		t.Fatal("incorrect description_expenses_uuid: ", errUUID)
	}
	if dataRespGet.DescriptionExpenseUUID != descriptionExpenseUUID {
		t.Fatalf("expected description-expenses_uuid %s got %s", descriptionExpenseUUID, dataRespGet.DescriptionExpenseUUID)
	}
	if _, errUUID := uuid.Parse(dataRespGet.ExpenseUUID); errUUID != nil {
		t.Fatal("incorrect expenses_uuid: ", errUUID)
	}
	if dataRespGet.ExpenseUUID != expenseUUID {
		t.Fatalf("expected expenses_uuid %s got %s", expenseUUID, dataRespGet.ExpenseUUID)
	}
}
func TestListDescriptionExpenseSuccessful(t *testing.T) {
	const (
		budgetUUID = "1a2b3c4d-5e6f-47a8-b9c0-1d2e3f4a5b6c"
		userUUID   = "4a5b6c7d-8e9f-40a1-b2c3-d4e5f6a7b8c9"
	)
	conf, _, router := App()
	accessToken := shared_testing.CreateTestAccessToken(userUUID, conf.Signature, t)
	testServer := httptest.NewServer(router)
	defer testServer.Close()
	dataQuery, errReadFileSql := os.ReadFile("load-mock-budget-data.sql")
	if errReadFileSql != nil {
		t.Fatal("failed to read file sql: ", errReadFileSql)
	}
	shared_testing.RefreshUserTestData(dataQuery, []string{"budgets", "expenses", "description_expenses"}, t)
	requestList, errReqList := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/expense/"+budgetUUID, nil)
	if errReqList != nil {
		t.Fatal("failed to prepare request: ", errReqList)
	}
	requestList.Header.Set("Authorization", "Bearer "+accessToken)
	respGet, errRespGet := http.DefaultClient.Do(requestList)
	if errRespGet != nil {
		t.Fatal("failed to get response: ", errRespGet)
	}
	dataRespList := shared_testing.HelperHandleResponse[[]model.DescriptionExpenses](respGet, http.StatusOK, t)
	if len(dataRespList) != 2 {
		t.Fatalf("expected len list budget %d got %d", 2, len(dataRespList))
	}
}
