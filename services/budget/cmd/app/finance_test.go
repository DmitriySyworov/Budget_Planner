package main

import (
	"app/budget-planner/internal/finance"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestFinanceSuccessful(t *testing.T) {
	const (
		budgetUUID  = "7c4e5186-b4d2-4ef8-a51b-0294cd93eb5a"
		expenseUUID = "3b8d4e9a-7c21-4f51-863a-2a1d4f6c90bc"
		userUUID    = "f20a1bc3-5819-4a92-bdde-62fa9c671b2e"
	)
	appVariable := App()
	accessToken := shtesting.CreateTestAccessToken(userUUID, appVariable.Conf.Signature, t)
	testServer := httptest.NewServer(appVariable.HandlerApp)
	defer testServer.Close()
	dataQuery, errReadFileSql := os.ReadFile("load-mock-budget-data.sql")
	if errReadFileSql != nil {
		t.Fatal("failed to read file sql: ", errReadFileSql)
	}
	shtesting.RefreshUserTestData(dataQuery, []string{"budgets", "expenses", "description_expenses"}, t)
	requestGet, errReqGet := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/finance/"+budgetUUID+"/"+expenseUUID, nil)
	if errReqGet != nil {
		t.Fatal("failed to prepare request: ", errReqGet)
	}
	requestGet.Header.Set("Authorization", "Bearer "+accessToken)
	respRemove, errRespGet := http.DefaultClient.Do(requestGet)
	if errRespGet != nil {
		t.Fatal("failed to get response: ", errRespGet)
	}
	dataRespGet := shtesting.HelperHandleResponse[finance.Finance](respRemove, http.StatusOK, t)
	t.Log(dataRespGet.Budget, dataRespGet.Expenses, dataRespGet.ExpensesPercent)
}
