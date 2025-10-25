package metricshandlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"socialpredict/models/modelstesting"
	"socialpredict/util"
)

func TestGetSystemMetricsHandler_Success(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	orig := util.DB
	util.DB = db
	t.Cleanup(func() {
		util.DB = orig
	})

	_, _ = modelstesting.UseStandardTestEconomics(t)

	user := modelstesting.GenerateUser("alice", 0)
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	req := httptest.NewRequest("GET", "/v0/system/metrics", nil)
	rec := httptest.NewRecorder()

	GetSystemMetricsHandler(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload["moneyCreated"] == nil {
		t.Fatalf("expected moneyCreated section in response: %+v", payload)
	}
}

func TestGetSystemMetricsHandler_Error(t *testing.T) {
	orig := util.DB
	util.DB = nil
	defer func() { util.DB = orig }()

	req := httptest.NewRequest("GET", "/v0/system/metrics", nil)
	rec := httptest.NewRecorder()

	GetSystemMetricsHandler(rec, req)

	if rec.Code != 500 {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}
