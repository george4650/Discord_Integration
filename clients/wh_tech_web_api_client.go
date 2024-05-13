package clients

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"wh-hard01.kol.wb.ru/wh-tech/wh-tech-back/wh_tech_discord_api/internal/models"
	client "wh-hard01.kol.wb.ru/wh_core/gocore_http"
	"wh-hard01.kol.wb.ru/wh_core/gocore_service_configs/configs"
)

type WhTechWebApiClient struct {
	fastHttpClient *client.HttpClient
}

func NewWhTechWebApiClient() *WhTechWebApiClient {
	return &WhTechWebApiClient{}
}

func (r *WhTechWebApiClient) Configure(ctx context.Context, config configs.Config) {
	r.fastHttpClient = client.NewHttpClient()
	r.fastHttpClient.InitHttpClient("wh_tech_web_api_client")(ctx, config)
}

func (r *WhTechWebApiClient) GetUserDataByEmployeeID(employeeID string) (string, error) {
	const api = "api/employeetags/get_by_employee_id/v2"

	params := map[string]string{
		"employee_id": employeeID,
	}

	resp, err := r.fastHttpClient.HTTPRequest(nil, http.MethodGet, nil, api, params)
	if err != nil {
		return "", r.fastHttpClient.GenerateError(resp, err, api)
	}

	if resp.StatusCode() == http.StatusOK {
		employeeInfo := models.EmployeeInfo{}
		err = jsoniter.Unmarshal(resp.Body(), &employeeInfo)
		if err != nil {
			return "", fmt.Errorf("[GetUserDataByEmployeeID] Unmarshal body error - %w", err)
		}
		return employeeInfo.Data[0].EmployeeName, nil
	}

	return "", r.fastHttpClient.GenerateError(resp, err, api)
}
