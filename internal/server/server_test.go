package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthRoute(t *testing.T) {
	rec := httptest.NewRecorder()
	New().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestProfileRoute(t *testing.T) {
	body := `{"case":{"rock":{"swc":0.2,"sor":0.2,"porosity":0.25},"relperm":{"krw0":0.4,"kro0":0.9,"nw":2.0,"no":2.0},"fluid":{"mu_w":1.0,"mu_o":2.0},"injection":{"sw_inj":0.8}}}`
	rec := httptest.NewRecorder()
	New().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/profile", strings.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestFractionalRoute(t *testing.T) {
	body := `{"case":{"rock":{"swc":0.2,"sor":0.2},"relperm":{"krw0":0.4,"kro0":0.9,"nw":2.0,"no":2.0},"fluid":{"mu_w":1.0,"mu_o":2.0},"injection":{"sw_inj":0.8}},"grid":11}`
	rec := httptest.NewRecorder()
	New().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/fractional", strings.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSweepRoute(t *testing.T) {
	body := `{"case":{"rock":{"swc":0.2,"sor":0.2},"relperm":{"krw0":0.4,"kro0":0.9,"nw":2.0,"no":2.0},"fluid":{"mu_w":1.0,"mu_o":2.0},"injection":{"sw_inj":0.8}},"param":"mu_w","from":0.5,"to":2.0,"steps":4}`
	rec := httptest.NewRecorder()
	New().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/sweep", strings.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}
