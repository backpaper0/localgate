package portal_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"localgate/internal/portal"
)

func TestHandlerReturns200(t *testing.T) {
	h := portal.NewHandler(2)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerContentTypeIsHTML(t *testing.T) {
	h := portal.NewHandler(2)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("expected Content-Type text/html, got %q", ct)
	}
}

func TestHandlerBodyIsNotEmpty(t *testing.T) {
	h := portal.NewHandler(2)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	if w.Body.Len() == 0 {
		t.Error("expected non-empty HTML body")
	}
}

func TestHandlerInjectsRefreshIntervalMs(t *testing.T) {
	h := portal.NewHandler(5)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	body := w.Body.String()
	if !strings.Contains(body, "5000") {
		t.Errorf("expected body to contain 5000 (5sec * 1000ms), got body len %d", len(body))
	}
}
