// Package portal は管理ポータルのHTMLページを配信するハンドラを提供する。
package portal

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed portal.html
var portalFS embed.FS

// portalData はHTMLテンプレートに渡すデータ
type portalData struct {
	RefreshIntervalMs int
}

// Handler は GET / に対してポータルHTMLを返すハンドラ
type Handler struct {
	tmpl              *template.Template
	refreshIntervalMs int
}

// NewHandler は新しいポータルハンドラを返す。
// refreshIntervalSec はポーリング間隔（秒）で、1以上であること。
// テンプレートのパースに失敗した場合はパニックする。
func NewHandler(refreshIntervalSec int) *Handler {
	tmpl := template.Must(template.ParseFS(portalFS, "portal.html"))
	return &Handler{
		tmpl:              tmpl,
		refreshIntervalMs: refreshIntervalSec * 1000,
	}
}

// ServeHTTP はポータルHTMLをレンダリングして返す。
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	h.tmpl.Execute(w, portalData{RefreshIntervalMs: h.refreshIntervalMs}) //nolint:errcheck
}
