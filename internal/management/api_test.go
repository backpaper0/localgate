package management_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"localgate/internal/management"
	"localgate/internal/registry"
)

func newAPI() (*management.API, registry.ServiceRegistry) {
	reg := registry.NewServiceRegistry()
	api := management.NewAPI(reg, 2)
	return api, reg
}

func TestPortalReturnsHTML(t *testing.T) {
	api, _ := newAPI()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	api.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("expected Content-Type text/html, got %q", ct)
	}

	if w.Body.Len() == 0 {
		t.Error("expected non-empty HTML body")
	}
}

func TestListServicesStillReturnsJSON(t *testing.T) {
	api, reg := newAPI()
	reg.Register("svc", "localhost:8080", false)

	r := httptest.NewRequest(http.MethodGet, "/services", nil)
	w := httptest.NewRecorder()

	api.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

func TestRegisterService(t *testing.T) {
	api, reg := newAPI()

	body := `{"name":"foo","target":"localhost:3000"}`
	r := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	api.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	_, found := reg.Lookup("foo")
	if !found {
		t.Error("expected service to be registered")
	}
}

func TestRegisterServiceMissingFields(t *testing.T) {
	api, _ := newAPI()

	body := `{"name":"foo"}`
	r := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	api.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegisterServiceInvalidJSON(t *testing.T) {
	api, _ := newAPI()

	r := httptest.NewRequest(http.MethodPost, "/services", strings.NewReader("invalid"))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	api.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDeregisterService(t *testing.T) {
	api, reg := newAPI()
	reg.Register("foo", "localhost:3000", false)

	r := httptest.NewRequest(http.MethodDelete, "/services/foo", nil)
	w := httptest.NewRecorder()

	api.ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}

	_, found := reg.Lookup("foo")
	if found {
		t.Error("expected service to be deregistered")
	}
}

func TestDeregisterServiceNotFound(t *testing.T) {
	api, _ := newAPI()

	r := httptest.NewRequest(http.MethodDelete, "/services/nonexistent", nil)
	w := httptest.NewRecorder()

	api.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestListServices(t *testing.T) {
	api, reg := newAPI()
	reg.Register("foo", "localhost:3000", false)
	reg.Register("bar", "localhost:4000", false)

	r := httptest.NewRequest(http.MethodGet, "/services", nil)
	w := httptest.NewRecorder()

	api.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp management.ListServicesResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(resp.Services))
	}
}

func TestListServicesEmpty(t *testing.T) {
	api, _ := newAPI()

	r := httptest.NewRequest(http.MethodGet, "/services", nil)
	w := httptest.NewRecorder()

	api.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp management.ListServicesResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Services == nil || len(resp.Services) != 0 {
		t.Errorf("expected empty services array, got %v", resp.Services)
	}
}
