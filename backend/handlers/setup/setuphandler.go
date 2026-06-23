package setup

import (
	"encoding/json"
	"net/http"

	"socialpredict/handlers"
	configsvc "socialpredict/internal/service/config"
)

func GetSetupHandler(configService configsvc.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if configService == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(configService.Economics())
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}

type frontendChartsResponse struct {
	SigFigs int `json:"sigFigs"`
}

type frontendGameResponse struct {
	Mode string `json:"mode"`
}

type frontendMarketGroupResponse struct {
	MultipleChoiceBinary frontendMultipleChoiceBinaryResponse `json:"multipleChoiceBinary"`
}

type frontendMultipleChoiceBinaryResponse struct {
	AddAnswerCost             int64 `json:"addAnswerCost"`
	SoftAnswerReviewThreshold int   `json:"softAnswerReviewThreshold"`
	HardAnswerSafetyCap       int   `json:"hardAnswerSafetyCap"`
}

type frontendConfigResponse struct {
	Charts       frontendChartsResponse      `json:"charts"`
	Game         frontendGameResponse        `json:"game"`
	MarketGroups frontendMarketGroupResponse `json:"marketGroups"`
}

func GetFrontendSetupHandler(configService configsvc.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if configService == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		economics := configService.Economics()
		response := frontendConfigResponse{
			Charts: frontendChartsResponse{
				SigFigs: configService.ChartSigFigs(),
			},
			Game: frontendGameResponse{
				Mode: configService.Game().Mode,
			},
			MarketGroups: frontendMarketGroupResponse{
				MultipleChoiceBinary: frontendMultipleChoiceBinaryResponse{
					AddAnswerCost:             economics.MarketIncentives.MultipleChoiceBinary.AddAnswerCost,
					SoftAnswerReviewThreshold: economics.MarketIncentives.MultipleChoiceBinary.SoftAnswerReviewThreshold,
					HardAnswerSafetyCap:       economics.MarketIncentives.MultipleChoiceBinary.HardAnswerSafetyCap,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
	}
}
