package main

import (
	"app/budget-planner/internal/budget"
	"app/budget-planner/internal/model"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"shared/shared_common"
	"shared/shared_testing"
	"testing"

	"github.com/google/uuid"
)

func TestCreateBudgetSuccessful(t *testing.T) {
	appVariable := App()
	accessToken := shared_testing.CreateTestAccessToken("2e4b3c1d-8f9a-4c2b-b5e1-d3a7f8c9e0b2", appVariable.Conf.Signature, t)
	testServer := httptest.NewServer(appVariable.HandlerApp)
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

var CaseDataUpdateBudget = []struct {
	Name string
	budget.RequestUpdateBudget
}{
	{Name: "test all data update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Amount: "1234.76", Start: "2027-06-01", Finish: "2027-09-23", Description: "new_Update"}},
	{Name: "test amount, start, finish update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Amount: "1234.76", Start: "2027-06-01", Finish: "2027-09-23"}},
	{Name: "test amount, start, description update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Amount: "1234.76", Start: "2027-06-01", Description: "new_Update"}},
	{Name: "test amount, finish, description update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Amount: "1234.76", Finish: "2027-09-23", Description: "new_Update"}},
	{Name: "test finish start, description update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Start: "2027-06-01", Finish: "2027-09-23", Description: "new_Update"}},
	{Name: "test amount, start update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Amount: "1234.76", Start: "2027-06-01"}},
	{Name: "test description, finish update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Finish: "2027-09-23", Description: "new_Update"}},
	{Name: "test amount, finish update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Amount: "1234.76", Finish: "2027-09-23"}},
	{Name: "test amount, description update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Amount: "1234.76", Description: "new_Update"}},
	{Name: "test start, finish update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Start: "2027-06-01", Finish: "2027-09-23"}},
	{Name: "test start, description update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Start: "2027-06-01", Description: "new_Update"}},
	{Name: "test amount update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Amount: "1234.76"}},
	{Name: "test start update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Start: "2027-06-01"}},
	{Name: "test finish update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Finish: "2027-09-23"}},
	{Name: "test description update budget - ", RequestUpdateBudget: budget.RequestUpdateBudget{Description: "new_Update"}},
}

func TestUpdateBudgetSuccessful(t *testing.T) {
	appVariable := App()
	testServer := httptest.NewServer(appVariable.HandlerApp)
	defer testServer.Close()
	dataQuery, errReadFileSql := os.ReadFile("load-mock-budget-data.sql")
	if errReadFileSql != nil {
		t.Fatal("failed to read file sql: ", errReadFileSql)
	}
	shared_testing.RefreshUserTestData(dataQuery, []string{"budgets", "expenses", "description_expenses"}, t)
	const (
		budgetUpdateUUID = "859c7a21-dc20-410a-ba54-2c11fb6db2a8"
		userUpdateUUID   = "1b272de3-9827-4c47-8a60-2da8e80556f8"
	)
	accessToken := shared_testing.CreateTestAccessToken(userUpdateUUID, appVariable.Conf.Signature, t)

	for _, test := range CaseDataUpdateBudget {
		data, errMarshalUpdate := json.Marshal(test)
		if errMarshalUpdate != nil {
			t.Fatal(test.Name+"failed to prepare request: ", errMarshalUpdate)
		}
		requestUpdate, errReqUpdate := http.NewRequest(http.MethodPatch, testServer.URL+"/api/v1/budget/"+budgetUpdateUUID, bytes.NewBuffer(data))
		if errReqUpdate != nil {
			t.Fatal(test.Name+"failed to prepare request: ", errReqUpdate)
		}
		requestUpdate.Header.Set("Authorization", "Bearer "+accessToken)
		respUpdate, errRespUpdate := http.DefaultClient.Do(requestUpdate)
		if errRespUpdate != nil {
			t.Fatal(test.Name+"failed to get response: ", errRespUpdate)
		}
		respData := shared_testing.HelperHandleResponse[model.Budgets](respUpdate, http.StatusOK, t)
		if respData.BudgetUUID != budgetUpdateUUID {
			t.Fatalf(test.Name+"budget_uuid: %s do not match %s", respData.BudgetUUID, budgetUpdateUUID)
		}
		if respData.UserUUID != userUpdateUUID {
			t.Fatalf(test.Name+"budget_uuid: %s do not match %s", respData.UserUUID, userUpdateUUID)
		}
	}
}
func TestGetBudgetSuccessful(t *testing.T) {
	appVariable := App()
	accessToken := shared_testing.CreateTestAccessToken("6e5f4a3b-2c1d-4e9f-8a7b-6c5d4e3f2a1b", appVariable.Conf.Signature, t)
	testServer := httptest.NewServer(appVariable.HandlerApp)
	defer testServer.Close()
	dataQuery, errReadFileSql := os.ReadFile("load-mock-budget-data.sql")
	if errReadFileSql != nil {
		t.Fatal("failed to read file sql: ", errReadFileSql)
	}
	budgetUUID := "0f1e2d3c-4b5a-4678-9abc-def012345678"
	shared_testing.RefreshUserTestData(dataQuery, []string{"budgets", "expenses", "description_expenses"}, t)
	requestRemove, errReqRemove := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/budget/"+budgetUUID, nil)
	if errReqRemove != nil {
		t.Fatal("failed to prepare request: ", errReqRemove)
	}
	requestRemove.Header.Set("Authorization", "Bearer "+accessToken)
	respRemove, errRespRemove := http.DefaultClient.Do(requestRemove)
	if errRespRemove != nil {
		t.Fatal("failed to get response: ", errRespRemove)
	}
	dataRespRemove := shared_testing.HelperHandleResponse[model.Budgets](respRemove, http.StatusOK, t)
	if _, errUUID := uuid.Parse(dataRespRemove.BudgetUUID); errUUID != nil {
		t.Fatal("incorrect budget_uuid: ", errUUID)
	}
}

var CaseDataRemoveBudget = []struct {
	Name       string
	BudgetUUID string
	UserUUID   string
	Type       string
}{
	{Name: "soft-delete budget - ", BudgetUUID: "b408d27c-19d8-42c4-8675-ae92166c8cf9", UserUUID: "3f9b95b0-e13e-4b44-bf46-75840e8fe52a", Type: shared_constant.TypeSoftDelete},
	{Name: "hard-delete budget - ", BudgetUUID: "671127d2-15d9-43a5-956c-5266f72204d0", UserUUID: "c6ccc482-9187-4baa-8925-0c60780627fe", Type: shared_constant.TypeHardDelete},
}

func TestRemoveBudgetSuccessful(t *testing.T) {
	appVariable := App()
	testServer := httptest.NewServer(appVariable.HandlerApp)
	defer testServer.Close()
	dataQuery, errReadFileSql := os.ReadFile("load-mock-budget-data.sql")
	if errReadFileSql != nil {
		t.Fatal("failed to read file sql: ", errReadFileSql)
	}
	db := shared_testing.RefreshUserTestData(dataQuery, []string{"budgets", "expenses", "description_expenses"}, t)
	for _, test := range CaseDataRemoveBudget {
		accessToken := shared_testing.CreateTestAccessToken(test.UserUUID, appVariable.Conf.Signature, t)
		requestGet, errReqGet := http.NewRequest(http.MethodDelete, testServer.URL+"/api/v1/budget/"+test.BudgetUUID+"?type="+test.Type, nil)
		if errReqGet != nil {
			t.Fatal(test.Name+"failed to prepare request: ", errReqGet)
		}
		requestGet.Header.Set("Authorization", "Bearer "+accessToken)
		respGet, errRespGet := http.DefaultClient.Do(requestGet)
		if errRespGet != nil {
			t.Fatal(test.Name+"failed to get response: ", errRespGet)
		}
		shared_testing.HelperHandleResponse[model.Budgets](respGet, http.StatusNoContent, t)
		budgets := &model.Budgets{}
		if test.Type == shared_constant.TypeSoftDelete {
			if db.Where("user_uuid = ? AND budget_uuid = ?", test.UserUUID, test.BudgetUUID).
				Take(budgets).Error == nil {
				t.Fatal(test.Name + "failed to remove user")
			}
		} else if test.Type == shared_constant.TypeHardDelete {
			if db.Unscoped().Where("user_uuid = ? AND budget_uuid = ?", test.UserUUID, test.BudgetUUID).
				Take(budgets).Error == nil {
				t.Fatal(test.Name + "failed to delete user")
			}
		}
	}
}

var CaseListBudgetData = []struct {
	Name                    string
	QueryParams             string
	ExpectedQuantityRecords int
}{
	{Name: "limit 2 offset 0 - ", QueryParams: fmt.Sprintf("?limit=%s&offset=%s", "2", "0"), ExpectedQuantityRecords: 2},
	{Name: "limit 1 offset 0 - ", QueryParams: fmt.Sprintf("?limit=%s&offset=%s", "1", "0"), ExpectedQuantityRecords: 1},
	{Name: "limit 1 offset 1 - ", QueryParams: fmt.Sprintf("?limit=%s&offset=%s", "1", "1"), ExpectedQuantityRecords: 1},
	{Name: "default values - ", QueryParams: "", ExpectedQuantityRecords: 2},
}

func TestListBudgetSuccessful(t *testing.T) {
	appVariable := App()
	const userListUUID = "c9b8a7d6-e5f4-4321-890a-bcdef1234567"
	accessToken := shared_testing.CreateTestAccessToken(userListUUID, appVariable.Conf.Signature, t)
	testServer := httptest.NewServer(appVariable.HandlerApp)
	defer testServer.Close()
	dataQuery, errReadFileSql := os.ReadFile("load-mock-budget-data.sql")
	if errReadFileSql != nil {
		t.Fatal("failed to read file sql: ", errReadFileSql)
	}
	shared_testing.RefreshUserTestData(dataQuery, []string{"budgets", "expenses", "description_expenses"}, t)
	for _, test := range CaseListBudgetData {
		requestGet, errReqGet := http.NewRequest(http.MethodGet, testServer.URL+"/api/v1/budget"+test.QueryParams, nil)
		if errReqGet != nil {
			t.Fatal(test.Name+"failed to prepare request: ", errReqGet)
		}
		requestGet.Header.Set("Authorization", "Bearer "+accessToken)
		respGet, errRespGet := http.DefaultClient.Do(requestGet)
		if errRespGet != nil {
			t.Fatal(test.Name+"failed to get response: ", errRespGet)
		}
		dataRespGet := shared_testing.HelperHandleResponse[[]model.Budgets](respGet, http.StatusOK, t)
		if len(dataRespGet) != test.ExpectedQuantityRecords {
			t.Fatalf(test.Name+"expected len list budget %d got %d", test.ExpectedQuantityRecords, len(dataRespGet))
		}
		if len(dataRespGet) >= 2 {
			userFirstUUID := dataRespGet[0].UserUUID
			userSecondUUID := dataRespGet[1].UserUUID
			if _, errUUID := uuid.Parse(userFirstUUID); errUUID != nil {
				t.Error(test.Name+"incorrect uuid: ", errUUID)
			}
			if _, errUUID := uuid.Parse(userSecondUUID); errUUID != nil {
				t.Error(test.Name+"incorrect uuid: ", errUUID)
			}
			if _, errUUID := uuid.Parse(dataRespGet[0].BudgetUUID); errUUID != nil {
				t.Error(test.Name+"incorrect uuid: ", errUUID)
			}
			if _, errUUID := uuid.Parse(dataRespGet[1].BudgetUUID); errUUID != nil {
				t.Error(test.Name+"incorrect uuid: ", errUUID)
			}
			if userFirstUUID != userListUUID || userSecondUUID != userListUUID {
				t.Fatal(test.Name + "user_uuid does not match in records")
			}
		}
	}
}
