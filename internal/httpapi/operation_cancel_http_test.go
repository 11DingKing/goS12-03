package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"batteryops/internal/domain"
)

// TestHTTPCancelledOperationExecuteRejected describes the expected REST
// behaviour after a grid-switching operation has been cancelled.
//
// Input: POST /api/operations creates a grid-connect operation, the supply
// station confirms it via POST /api/operations/{id}/confirm, and the dispatcher
// then cancels it via POST /api/operations/{id}/cancel.
//
// Expected output: the following POST /api/operations/{id}/execute must be
// refused with 409 Conflict, and GET /api/operations must still report the
// operation as cancelled without an execution timestamp.
func TestHTTPCancelledOperationExecuteRejected(t *testing.T) {
	server, _ := newTestServer()

	req := httptest.NewRequest("POST", "/api/operations",
		strings.NewReader(`{"type":"grid_connect","dispatcher_id":"dispatcher-1"}`))
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("initiate: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var op domain.Operation
	if err := json.NewDecoder(w.Body).Decode(&op); err != nil {
		t.Fatalf("decode: %v", err)
	}

	req = httptest.NewRequest("POST", "/api/operations/"+op.ID+"/confirm",
		strings.NewReader(`{"station_id":"station-1"}`))
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("confirm: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest("POST", "/api/operations/"+op.ID+"/cancel",
		strings.NewReader(`{"actor":"dispatcher-1"}`))
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("cancel: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest("POST", "/api/operations/"+op.ID+"/execute", nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("execute after cancel: expected 409, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest("GET", "/api/operations", nil)
	w = httptest.NewRecorder()
	server.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list operations: expected 200, got %d", w.Code)
	}
	var listed []domain.Operation
	if err := json.NewDecoder(w.Body).Decode(&listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(listed))
	}
	if listed[0].Status != domain.OperationCancelled {
		t.Fatalf("expected status %s, got %s", domain.OperationCancelled, listed[0].Status)
	}
	if listed[0].ExecutedAt != nil {
		t.Fatalf("expected no executed_at on a cancelled operation, got %v", listed[0].ExecutedAt)
	}
}
