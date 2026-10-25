package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewEngineRegistersNewRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e := NewEngine(Deps{
		Auth:       &AuthHandler{},
		Tunnel:     &TunnelHandler{},
		Domain:     &DomainHandler{},
		DomainPool: &DomainPoolHandler{},
		Admin:      &AdminHandler{},
		Session:    &SessionHandler{},
		Log:        &LogHandler{},
		Stats:      &StatsHandler{},
		Event:      &EventHandler{},
		JWTSecret:  "test",
	})
	want := map[string]bool{
		"GET /api/domains":             false,
		"POST /api/domains":            false,
		"PUT /api/domains/:id":         false,
		"DELETE /api/domains/:id":      false,
		"GET /api/domains/available":   false,
		"GET /api/admins":              false,
		"POST /api/admins":             false,
		"PUT /api/admins/:id":          false,
		"PUT /api/admins/:id/password": false,
		"PUT /api/admins/:id/domains":  false,
		"DELETE /api/admins/:id":       false,
	}
	for _, r := range e.Routes() {
		key := r.Method + " " + r.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for key, ok := range want {
		if !ok {
			t.Errorf("route not registered: %s", key)
		}
	}
}
