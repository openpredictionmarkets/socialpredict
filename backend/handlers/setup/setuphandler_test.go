package setup

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"socialpredict/setup"
	"strings"
	"testing"
)

// TODO Before submitting: Review these tests to ensure they make sense with the new config loader
func TestGetSetupHandler(t *testing.T) {
	tests := []struct {
		Name             string
		MockConfigLoader setup.EconConfigLoader
		ExpectedStatus   int
		ExpectedResponse string
		IsJSONResponse   bool
	}{
		{
			Name: "LoadProductionConfig",
			MockConfigLoader: func() *setup.EconomicConfig {
				return setup.MustLoadEconomicsConfig()
			},
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: `{
				"MarketCreation":{"InitialMarketProbability":0.5,"InitialMarketSubsidization":10,"InitialMarketYes":0,"InitialMarketNo":0},
				"MarketIncentives":{"CreateMarketCost":10,"TraderBonus":1},
				"User":{"InitialAccountBalance":0,"MaximumDebtAllowed":500},
				"Betting":{"MinimumBet":1,"BetFees":{"InitialBetFee":1,"EachBetFee":0,"SellSharesFee":0}}}`,
			IsJSONResponse: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// Replace the actual loader function with the mock
			loadEconomicsConfig := test.MockConfigLoader

			req, err := http.NewRequest("GET", "/setup", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(GetSetupHandler(loadEconomicsConfig))

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

// jsonEqual compares two JSON objects for equality
func jsonEqual(a, b map[string]interface{}) bool {
	return jsonString(a) == jsonString(b)
}

// jsonString converts a JSON object to a sorted string
func jsonString(v interface{}) string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}
