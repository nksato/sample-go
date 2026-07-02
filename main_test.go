package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestStore returns a store with predictable taxi positions for testing.
func newTestStore() *taxiStore {
	return &taxiStore{
		taxis: []Taxi{
			{ID: "A", Lat: 35.68, Lng: 139.76},
			{ID: "B", Lat: 35.69, Lng: 139.77},
			{ID: "C", Lat: 35.70, Lng: 139.78},
		},
	}
}

func TestGetTaxis(t *testing.T) {
	store := newTestStore()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.all())
	})

	req := httptest.NewRequest(http.MethodGet, "/api/taxis", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var taxis []Taxi
	if err := json.NewDecoder(rec.Body).Decode(&taxis); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if len(taxis) != 3 {
		t.Fatalf("expected 3 taxis, got %d", len(taxis))
	}
}

func TestGetTaxisMethodNotAllowed(t *testing.T) {
	store := newTestStore()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.all())
	})

	req := httptest.NewRequest(http.MethodPost, "/api/taxis", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestCallTaxi(t *testing.T) {
	store := newTestStore()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req CallRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		id := store.nearest(req.Lat, req.Lng)
		resp := CallResponse{
			Message: "タクシーを手配しました。タクシー " + id + " が向かっています。",
			TaxiID:  id,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	body, _ := json.Marshal(CallRequest{Lat: 35.68, Lng: 139.76})
	req := httptest.NewRequest(http.MethodPost, "/api/call", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp CallResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	// Taxi A is at (35.68, 139.76) – exactly the requested position, so it must win.
	if resp.TaxiID != "A" {
		t.Errorf("expected nearest taxi A, got %s", resp.TaxiID)
	}

	if resp.Message == "" {
		t.Error("expected non-empty message")
	}
}

func TestCallTaxiInvalidBody(t *testing.T) {
	store := newTestStore()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req CallRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		id := store.nearest(req.Lat, req.Lng)
		resp := CallResponse{
			Message: "タクシーを手配しました。タクシー " + id + " が向かっています。",
			TaxiID:  id,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/call", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestNearestTaxi(t *testing.T) {
	store := newTestStore()

	// Position closest to taxi B.
	id := store.nearest(35.69, 139.77)
	if id != "B" {
		t.Errorf("expected B, got %s", id)
	}
}

func TestTaxiMove(t *testing.T) {
	store := newTestStore()
	before := store.all()
	store.move()
	after := store.all()

	if len(before) != len(after) {
		t.Fatal("move changed the number of taxis")
	}
	// After a move, at least one coordinate should have changed.
	changed := false
	for i := range before {
		if before[i].Lat != after[i].Lat || before[i].Lng != after[i].Lng {
			changed = true
			break
		}
	}
	if !changed {
		t.Error("move did not change any taxi position")
	}
}
