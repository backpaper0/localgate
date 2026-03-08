package management

import (
	"encoding/json"
	"errors"
	"net/http"

	"localgate/internal/portal"
	"localgate/internal/registry"
)

// ListServicesResponse は GET /services のレスポンス形式
type ListServicesResponse struct {
	Services []registry.ServiceEntry `json:"services"`
}

type registerServiceRequest struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

// API は管理HTTPエンドポイントを処理するハンドラ
type API struct {
	registry registry.ServiceRegistry
	mux      *http.ServeMux
}

// NewAPI は新しい管理APIハンドラを返す。
// refreshIntervalSec はポータル画面のポーリング間隔（秒）。
func NewAPI(reg registry.ServiceRegistry, refreshIntervalSec int) *API {
	api := &API{registry: reg}
	mux := http.NewServeMux()
	mux.Handle("GET /", portal.NewHandler(refreshIntervalSec))
	mux.HandleFunc("POST /services", api.handleRegister)
	mux.HandleFunc("DELETE /services/{name}", api.handleDeregister)
	mux.HandleFunc("GET /services", api.handleList)
	api.mux = mux
	return api
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Target == "" {
		writeError(w, http.StatusBadRequest, "name and target are required")
		return
	}
	force := r.Header.Get("X-Force-Overwrite") == "true"
	if err := a.registry.Register(req.Name, req.Target, force); err != nil {
		if errors.Is(err, registry.ErrAlreadyExists) {
			existingTarget, _ := a.registry.Lookup(req.Name)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error":           "service already exists",
				"existing_target": existingTarget,
			})
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(registry.ServiceEntry{Name: req.Name, Target: req.Target})
}

func (a *API) handleDeregister(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := a.registry.Deregister(name); err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			writeError(w, http.StatusNotFound, "service not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleList(w http.ResponseWriter, r *http.Request) {
	entries := a.registry.List()
	if entries == nil {
		entries = []registry.ServiceEntry{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ListServicesResponse{Services: entries})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
