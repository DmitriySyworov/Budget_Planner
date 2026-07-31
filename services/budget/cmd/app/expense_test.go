package main

import (
	"app/budget-planner/internal/expense"
	"app/budget-planner/internal/model"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
)

var CaseCreateDescriptionExpenseData = []struct {
	Name       string
	BudgetUUID string
}{
	{Name: "update expense - ", BudgetUUID: "f47ac10b-58cc-4372-a567-0e02b2c3d479"},
	{Name: "create expense - ", BudgetUUID: "db0e2a3c-5014-416b-a8f2-8ef96bc752a7"},
}

func TestCreateExpenseSuccessful(t *testing.T) {
	const (
		userCreateUUID = "26f50b4d-389d-4de3-85f5-46bd8ecb1d03"
	)
	appVariable := App()
	accessToken := shtesting.CreateTestAccessToken(userCreateUUID, appVariable.Conf.Signature, t)
	testServer := httptest.NewServer(appVariable.HandlerApp)
	defer testServer.Close()
	dataQuery, errReadFileSql := os.ReadFile("load-mock-budget-data.sql")
	if errReadFileSql != nil {
		t.Fatal("failed to read file sql: ", errReadFileSql)
	}
	for _, test := range CaseCreateDescriptionExpenseData {
		shtesting.RefreshUserTestData(dataQuery, []string{"budgets", "expenses", "description_expenses"}, t)
		bodyCreateExpense := expense.RequestCreateDescriptionExpense{
			Category:    "sport",
			Expense:     "234.78",
			Description: "buy pills",
		}
		dataCreate, errMarshalCreate := json.Marshal(bodyCreateExpense)
		if errMarshalCreate != nil {
			t.Fatal(test.Name+"failed to prepare body: ", errMarshalCreate)
		}
		requestCreate, errReqCreate := http.NewRequest(http.MethodPost, testServer.URL+"/api/v1/description-expense/"+test.BudgetUUID, bytes.NewBuffer(dataCreate))
		if errReqCreate != nil {
			t.Fatal(test.Name+"failed to prepare request: ", errReqCreate)
		}
		requestCreate.Header.Set("Authorization", "Bearer "+accessToken)
		respCreate, errRespCreate := http.DefaultClient.Do(requestCreate)
		if errRespCreate != nil {
			t.Fatal(test.Name+"failed to get response: ", errRespCreate)
		}
		dataRespCreate := shtesting.HelperHandleResponse[expense.ResponseCreateAndUpdateExpense](respCreate, http.StatusCreated, t)
		if _, errUUID := uuid.Parse(dataRespCreate.DescriptionExpenseUUID); errUUID != nil {
			t.Fatal(test.Name+"incorrect description_expense_uuid: ", errUUID)
		}
		if _, errUUID := uuid.Parse(dataRespCreate.Expenses.ExpenseUUID); errUUID != nil {
			t.Fatal(test.Name+"incorrect expense_uuid: ", errUUID)
		}
		if bodyCreateExpense.Expense != dataRespCreate.Expenses.Sport {
			t.Fatalf("%sexpected amount %s got %s", test.Name, bodyCreateExpense.Expense, dataRespCreate.Expenses.Sport)
		}
	}
}

var CaseUpdateDescriptionExpenseData = []struct {
	Name string
	expense.RequestUpdateDescriptionExpense
}{
	{Name: "update all - ", RequestUpdateDescriptionExpense: expense.RequestUpdateDescriptionExpense{Category: "sport", Expense: "1230.00", Description: "new update expense"}},
	{Name: "update category and expense - ", RequestUpdateDescriptionExpense: expense.RequestUpdateDescriptionExpense{Category: "sport", Expense: "1230.00"}},
	{Name: "update category and description - ", RequestUpdateDescriptionExpense: expense.RequestUpdateDescriptionExpense{Category: "sport", Description: "new update expense"}},
	{Name: "update expense and description - ", RequestUpdateDescriptionExpense: expense.RequestUpdateDescriptionExpense{Expense: "1230.00", Description: "new update expense"}},
	{Name: "update expense - ", RequestUpdateDescriptionExpense: expense.RequestUpdateDescriptionExpense{Expense: "1230.00"}},
	{Name: "update Description - ", RequestUpdateDescriptionExpense: expense.RequestUpdateDescriptionExpense{Description: "new update expense"}},
	{Name: "update category - ", RequestUpdateDescriptionExpense: expense.RequestUpdateDescriptionExpense{Category: "sport"}},
}

func TestUpdateDescriptionExpenseSuccessful(t *testing.T) {
	const (
		budgetUUID             = "dd6049ad-8422-44af-aa6f-8de18f2f64d0"
		expenseUUID            = "28df9dc1-be18-4da5-9bd9-f9c3fc88f4ef"
		descriptionExpenseUUID = "a51052ab-9cf4-4b55-b46f-c1dcdeec21a2"
		userUUID               = "941865fb-cde7-456b-a25e-de8159b9bd14"
	)
	appVariable := App()
	testServer := httptest.NewServer(appVariable.HandlerApp)
	defer testServer.Close()
	dataQuery, errReadFileSql := os.ReadFile("load-mock-budget-data.sql")
	if errReadFileSql != nil {
		t.Fatal("failed to read file sql: ", errReadFileSql)
	}
	for _, test := range CaseUpdateDescriptionExpenseData {
		accessToken := shtesting.CreateTestAccessToken(userUUID, appVariable.Conf.Signature, t)
		shtesting.RefreshUserTestData(dataQuery, []string{"budgets", "expenses", "description_expenses"}, t)
		data, errMarshal := json.Marshal(test)
		if errMarshal != nil {
			t.Fatal(test.Name+"failed to prepare request: ", errMarshal)
		}
		requestGet, errReqGet := http.NewRequest(http.MethodPatch, testServer.URL+"/api/v1/description-expense/"+budgetUUID+"/"+descriptionExpenseUUID, bytes.NewBuffer(data))
		if errReqGet != nil {
			t.Fatal(test.Name+"failed to prepare request: ", errReqGet)
		}
		requestGet.Header.Set("Authorization", "Bearer "+accessToken)
		respRemove, errRespGet := http.DefaultClient.Do(requestGet)
		if errRespGet != nil {
			t.Fatal(test.Name+"failed to get response: ", errRespGet)
		}
		dataRespGet := shtesting.HelperHandleResponse[expense.ResponseCreateAndUpdateExpense](respRemove, http.StatusOK, t)
		if _, errUUID := uuid.Parse(dataRespGet.DescriptionExpenseUUID); errUUID != nil {
			t.Fatal(test.Name+"incorrect description_expenses_uuid: ", errUUID)
		}
		if dataRespGet.DescriptionExpenses.DescriptionExpenseUUID != descriptionExpenseUUID {
			t.Fatalf(test.Name+"expected description_expenses_uuid %s got %s", descriptionExpenseUUID, dataRespGet.DescriptionExpenses.DescriptionExpenseUUID)
		}
		if _, errUUID := uuid.Parse(dataRespGet.Expenses.ExpenseUUID); errUUID != nil {
			t.Fatal(test.Name+"incorrect expenses_uuid: ", errUUID)
		}
		if dataRespGet.Expenses.ExpenseUUID != expenseUUID {
			t.Fatalf(test.Name+"expected expenses_uuid %s got %s", expenseUUID, dataRespGet.Expenses.ExpenseUUID)
		}
		t.Log(dataRespGet.Expenses)
		t.Log(dataRespGet.DescriptionExpenses)
	}
}
func TestGetDescriptionExpenseSuccessful(t *testing.T) {
	const (
		budgetUUID             = "1a2b3c4d-5e6f-47a8-b9c0-1d2e3f4a5b6c"
		expenseUUID            = "9926d83a-4be4-4298-ba98-25081b29cc36"
		descriptionExpenseUUID = "5b8b9333-d922-4a00-bf86-53d368e734bc"
		userUUID               = "4a5b6c7d-8e9f-40a1-b2c3-d4e5f6a7b8c9"
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
	requestGet, errReqGet := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/description-expense/"+budgetUUID+"/"+descriptionExpenseUUID, nil)
	if errReqGet != nil {
		t.Fatal("failed to prepare request: ", errReqGet)
	}
	requestGet.Header.Set("Authorization", "Bearer "+accessToken)
	respRemove, errRespGet := http.DefaultClient.Do(requestGet)
	if errRespGet != nil {
		t.Fatal("failed to get response: ", errRespGet)
	}
	dataRespGet := shtesting.HelperHandleResponse[model.DescriptionExpenses](respRemove, http.StatusOK, t)
	if _, errUUID := uuid.Parse(dataRespGet.DescriptionExpenseUUID); errUUID != nil {
		t.Fatal("incorrect description_expenses_uuid: ", errUUID)
	}
	if dataRespGet.DescriptionExpenseUUID != descriptionExpenseUUID {
		t.Fatalf("expected description_expenses_uuid %s got %s", descriptionExpenseUUID, dataRespGet.DescriptionExpenseUUID)
	}
	if _, errUUID := uuid.Parse(dataRespGet.ExpenseUUID); errUUID != nil {
		t.Fatal("incorrect expenses_uuid: ", errUUID)
	}
	if dataRespGet.ExpenseUUID != expenseUUID {
		t.Fatalf("expected expenses_uuid %s got %s", expenseUUID, dataRespGet.ExpenseUUID)
	}
}
func TestGetExpenseSuccessful(t *testing.T) {
	const (
		budgetUUID  = "1a2b3c4d-5e6f-47a8-b9c0-1d2e3f4a5b6c"
		expenseUUID = "9926d83a-4be4-4298-ba98-25081b29cc36"
		userUUID    = "4a5b6c7d-8e9f-40a1-b2c3-d4e5f6a7b8c9"
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
	requestGet, errReqGet := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/expense/"+budgetUUID, nil)
	if errReqGet != nil {
		t.Fatal("failed to prepare request: ", errReqGet)
	}
	requestGet.Header.Set("Authorization", "Bearer "+accessToken)
	respRemove, errRespGet := http.DefaultClient.Do(requestGet)
	if errRespGet != nil {
		t.Fatal("failed to get response: ", errRespGet)
	}
	dataRespGet := shtesting.HelperHandleResponse[model.Expenses](respRemove, http.StatusOK, t)
	if _, errUUID := uuid.Parse(dataRespGet.ExpenseUUID); errUUID != nil {
		t.Fatal("incorrect expenses_uuid: ", errUUID)
	}
	if dataRespGet.ExpenseUUID != expenseUUID {
		t.Fatalf("expected expenses_uuid %s got %s", expenseUUID, dataRespGet.ExpenseUUID)
	}
}
func TestDeleteDescriptionExpenseSuccessful(t *testing.T) {
	const (
		budgetUUID             = "e7d4a215-ca89-4ec4-bcde-51dcbaf642bd"
		expenseUUID            = "82baf21b-beec-48cb-8db1-cb5d0e2c39d8"
		descriptionExpenseUUID = "3047d8b5-0c14-41cf-a0e2-63b7ad6ab6cd"
		userUUID               = "6e917dcb-94ad-4cd6-b3de-857e2c943dfc"
	)
	appVariable := App()
	accessToken := shtesting.CreateTestAccessToken(userUUID, appVariable.Conf.Signature, t)
	testServer := httptest.NewServer(appVariable.HandlerApp)
	defer testServer.Close()
	dataQuery, errReadFileSql := os.ReadFile("load-mock-budget-data.sql")
	if errReadFileSql != nil {
		t.Fatal("failed to read file sql: ", errReadFileSql)
	}
	db := shtesting.RefreshUserTestData(dataQuery, []string{"budgets", "expenses", "description_expenses"}, t)
	requestGet, errReqGet := http.NewRequest(http.MethodDelete, testServer.URL+"/api/v1/description-expense/"+budgetUUID+"/"+descriptionExpenseUUID, nil)
	if errReqGet != nil {
		t.Fatal("failed to prepare request: ", errReqGet)
	}
	requestGet.Header.Set("Authorization", "Bearer "+accessToken)
	respRemove, errRespGet := http.DefaultClient.Do(requestGet)
	if errRespGet != nil {
		t.Fatal("failed to get response: ", errRespGet)
	}
	shtesting.HelperHandleResponse[struct{}](respRemove, http.StatusNoContent, t)
	if db.Where("expense_uuid = ? AND description_expense_uuid = ?", expenseUUID, descriptionExpenseUUID).
		Take(&model.DescriptionExpenses{}).Error == nil {
		t.Fatal("failed to delete description_expenses")
	}
	var amount string
	if errCheckExpense := db.Raw("SELECT sport FROM expenses WHERE expense_uuid = ? AND budget_uuid = ?", expenseUUID, budgetUUID).
		Scan(&amount).Error; errCheckExpense != nil {
		t.Fatal("failed to check result expenses: ", errCheckExpense)
	}
	if amount != "1.00" {
		t.Fatalf("expected 1.00 got %s amount", amount)
	}
}

var CaseListDescriptionExpenseData = []struct {
	Name                    string
	QueryParams             string
	ExpectedQuantityRecords int
}{
	{Name: "limit 2 offset 0 - ", QueryParams: fmt.Sprintf("?limit=%s&offset=%s", "2", "0"), ExpectedQuantityRecords: 2},
	{Name: "limit 1 offset 0 - ", QueryParams: fmt.Sprintf("?limit=%s&offset=%s", "1", "0"), ExpectedQuantityRecords: 1},
	{Name: "limit 1 offset 1 - ", QueryParams: fmt.Sprintf("?limit=%s&offset=%s", "1", "1"), ExpectedQuantityRecords: 1},
	{Name: "default values - ", QueryParams: "", ExpectedQuantityRecords: 2},
}

func TestListDescriptionExpenseSuccessful(t *testing.T) {
	const (
		budgetUUID = "1a2b3c4d-5e6f-47a8-b9c0-1d2e3f4a5b6c"
		userUUID   = "4a5b6c7d-8e9f-40a1-b2c3-d4e5f6a7b8c9"
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
	for _, test := range CaseListDescriptionExpenseData {
		requestList, errReqList := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/description-expense/"+budgetUUID+test.QueryParams, nil)
		if errReqList != nil {
			t.Fatal(test.Name+"failed to prepare request: ", errReqList)
		}
		requestList.Header.Set("Authorization", "Bearer "+accessToken)
		respGet, errRespGet := http.DefaultClient.Do(requestList)
		if errRespGet != nil {
			t.Fatal(test.Name+"failed to get response: ", errRespGet)
		}
		dataRespList := shtesting.HelperHandleResponse[[]model.DescriptionExpenses](respGet, http.StatusOK, t)
		if len(dataRespList) != test.ExpectedQuantityRecords {
			t.Fatalf(test.Name+"expected len list budget %d got %d", test.ExpectedQuantityRecords, len(dataRespList))
		}
	}
}
