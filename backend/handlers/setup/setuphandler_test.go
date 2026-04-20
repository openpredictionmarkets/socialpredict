package setup

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	configsvc "socialpredict/internal/service/config"
	"socialpredict/models/modelstesting"
	"socialpredict/setup"
)

func TestGetSetupHandler(t *testing.T) {
	tests := []struct {
		Name             string
		ConfigService    configsvc.Service
		ExpectedStatus   int
		ExpectedResponse string
		IsJSONResponse   bool
	}{
		{
			Name:           "successful load",
			ConfigService:  configsvc.NewStaticService(loadSetupConfig(t)),
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: `{
				"MarketCreation":{"InitialMarketProbability":0.5,"InitialMarketSubsidization":10,"InitialMarketYes":0,"InitialMarketNo":0,"MinimumFutureHours":1},
				"MarketIncentives":{"CreateMarketCost":10,"TraderBonus":1},
				"User":{"InitialAccountBalance":0,"MaximumDebtAllowed":500},
				"Betting":{"MinimumBet":1,"MaxDustPerSale":2,"BetFees":{"InitialBetFee":1,"BuySharesFee":0,"SellSharesFee":0}}}`,
			IsJSONResponse: true,
		}, {
			Name:             "missing config service",
			ConfigService:    nil,
			ExpectedStatus:   http.StatusInternalServerError,
			ExpectedResponse: "Failed to load economic config",
			IsJSONResponse:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/setup", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(GetSetupHandler(test.ConfigService))

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != test.ExpectedStatus {
				t.Errorf("%s: handler returned wrong status code: got %v want %v",
					test.Name, status, test.ExpectedStatus)
			}

			if test.IsJSONResponse {
				var expectedjson, actualjson map[string]interface{}
				if err := json.Unmarshal([]byte(test.ExpectedResponse), &expectedjson); err != nil {
					t.Fatalf("%s: error parsing expected response JSON: %v", test.Name, err)
				}
				if err := json.Unmarshal(rr.Body.Bytes(), &actualjson); err != nil {
					t.Fatalf("%s: error parsing actual response JSON: %v", test.Name, err)
				}

				if !jsonEqual(expectedjson, actualjson) {
					t.Errorf("%s: handler returned unexpected body: got %v want %v",
						test.Name, rr.Body.String(), test.ExpectedResponse)
				}
			} else {
				if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(test.ExpectedResponse) {
					t.Errorf("%s: handler returned unexpected body: got %v want %v",
						test.Name, rr.Body.String(), test.ExpectedResponse)
				}
			}
		})
	}
}

func TestGetFrontendSetupHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/setup/frontend", nil)
	rr := httptest.NewRecorder()

	config := modelstesting.GenerateEconomicConfig()
	config.Frontend.Charts.SigFigs = 1

	handler := http.HandlerFunc(GetFrontendSetupHandler(configsvc.NewStaticService(config)))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]map[string]int
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got := response["charts"]["sigFigs"]; got != 2 {
		t.Fatalf("expected clamped sig figs 2, got %d", got)
	}
}

func loadSetupConfig(t *testing.T) *setup.EconomicConfig {
	t.Helper()

	config := modelstesting.GenerateEconomicConfig()
	config.Economics.MarketCreation.MinimumFutureHours = 1
	return config
}

// jsonEqual compares two JSON objects for equality
func jsonEqual(a, b map[string]interface{}) bool {
	return jsonString(a) == jsonString(b)
}

// jsonString converts a JSON object to a sorted string
func jsonString(v interface{}) string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}
