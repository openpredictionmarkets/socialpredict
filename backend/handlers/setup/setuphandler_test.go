package setup

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"socialpredict/handlers"
	configsvc "socialpredict/internal/service/config"
	"socialpredict/models/modelstesting"
	"socialpredict/setup"
)

func TestGetSetupHandler(t *testing.T) {
	ownedConfig := configsvc.FromSetup(loadSetupConfig(t))

	tests := []struct {
		Name             string
		ConfigService    configsvc.Service
		ExpectedStatus   int
		ExpectedResponse string
		IsJSONResponse   bool
	}{
		{
			Name:           "successful load",
			ConfigService:  economicsOnlyConfigService{economics: ownedConfig.Economics},
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: `{
				"marketcreation":{"initialMarketProbability":0.5,"initialMarketSubsidization":10,"initialMarketYes":0,"initialMarketNo":0,"minimumFutureHours":1},
				"marketincentives":{"createMarketCost":10,"traderBonus":1},
				"user":{"initialAccountBalance":0,"maximumDebtAllowed":500},
				"betting":{"minimumBet":1,"maxDustPerSale":1,"betFees":{"initialBetFee":1,"buySharesFee":0,"sellSharesFee":0}}}`,
			IsJSONResponse: true,
		}, {
			Name:             "missing config service",
			ConfigService:    nil,
			ExpectedStatus:   http.StatusInternalServerError,
			ExpectedResponse: string(handlers.ReasonInternalError),
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
				var response handlers.FailureEnvelope
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("%s: error parsing failure response JSON: %v", test.Name, err)
				}
				if response.OK || response.Reason != test.ExpectedResponse {
					t.Errorf("%s: handler returned unexpected failure: got %+v want reason %v",
						test.Name, response, test.ExpectedResponse)
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
	config.Game.Mode = configsvc.GameModeModerator

	handler := http.HandlerFunc(GetFrontendSetupHandler(configsvc.NewStaticService(config)))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response struct {
		Charts struct {
			SigFigs int `json:"sigFigs"`
		} `json:"charts"`
		Game struct {
			Mode string `json:"mode"`
		} `json:"game"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got := response.Charts.SigFigs; got != 2 {
		t.Fatalf("expected clamped sig figs 2, got %d", got)
	}
	if got := response.Game.Mode; got != configsvc.GameModeModerator {
		t.Fatalf("expected game mode moderator, got %q", got)
	}
}

func TestGetFrontendSetupHandlerUsesChartSigFigsAccessor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/setup/frontend", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetFrontendSetupHandler(chartSigFigsOnlyConfigService{chartSigFigs: 9}))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response struct {
		Charts struct {
			SigFigs int `json:"sigFigs"`
		} `json:"charts"`
		Game struct {
			Mode string `json:"mode"`
		} `json:"game"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got := response.Charts.SigFigs; got != 9 {
		t.Fatalf("expected sig figs 9, got %d", got)
	}
	if got := response.Game.Mode; got != configsvc.GameModeModerator {
		t.Fatalf("expected game mode moderator, got %q", got)
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

type economicsOnlyConfigService struct {
	economics configsvc.Economics
}

func (s economicsOnlyConfigService) Current() *configsvc.AppConfig {
	panic("Current should not be called")
}

func (s economicsOnlyConfigService) Economics() configsvc.Economics {
	return s.economics
}

func (economicsOnlyConfigService) Frontend() configsvc.Frontend {
	panic("Frontend should not be called")
}

func (economicsOnlyConfigService) Game() configsvc.Game {
	panic("Game should not be called")
}

func (economicsOnlyConfigService) ChartSigFigs() int {
	panic("ChartSigFigs should not be called")
}

type chartSigFigsOnlyConfigService struct {
	chartSigFigs int
}

func (chartSigFigsOnlyConfigService) Current() *configsvc.AppConfig {
	panic("Current should not be called")
}

func (chartSigFigsOnlyConfigService) Economics() configsvc.Economics {
	panic("Economics should not be called")
}

func (chartSigFigsOnlyConfigService) Frontend() configsvc.Frontend {
	panic("Frontend should not be called")
}

func (chartSigFigsOnlyConfigService) Game() configsvc.Game {
	return configsvc.Game{Mode: configsvc.GameModeModerator}
}

func (s chartSigFigsOnlyConfigService) ChartSigFigs() int {
	return s.chartSigFigs
}
