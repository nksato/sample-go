package main

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

// Taxi represents a taxi with a unique ID and its current GPS position.
type Taxi struct {
	ID  string  `json:"id"`
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// CallRequest is the body sent when a user requests a taxi.
type CallRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// CallResponse is returned after a taxi-call request.
type CallResponse struct {
	Message string `json:"message"`
	TaxiID  string `json:"taxiId"`
}

// taxiStore holds a simulated fleet of taxis that drift randomly over time.
type taxiStore struct {
	mu    sync.RWMutex
	taxis []Taxi
}

// newTaxiStore creates a fleet centred at the given coordinates.
func newTaxiStore(centerLat, centerLng float64) *taxiStore {
	const count = 5
	taxis := make([]Taxi, count)
	for i := 0; i < count; i++ {
		taxis[i] = Taxi{
			ID:  taxiID(i),
			Lat: centerLat + (rand.Float64()-0.5)*0.02,
			Lng: centerLng + (rand.Float64()-0.5)*0.02,
		}
	}
	return &taxiStore{taxis: taxis}
}

func taxiID(i int) string {
	return string(rune('A' + i))
}

// move drifts each taxi a tiny random amount to simulate real movement.
func (s *taxiStore) move() {
	const step = 0.0003
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.taxis {
		s.taxis[i].Lat += (rand.Float64()-0.5) * step
		s.taxis[i].Lng += (rand.Float64()-0.5) * step
	}
}

// all returns a snapshot of all taxis.
func (s *taxiStore) all() []Taxi {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Taxi, len(s.taxis))
	copy(out, s.taxis)
	return out
}

// nearest returns the ID of the closest taxi to (lat, lng).
func (s *taxiStore) nearest(lat, lng float64) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	best := ""
	bestDist := math.MaxFloat64
	for _, t := range s.taxis {
		d := math.Hypot(t.Lat-lat, t.Lng-lng)
		if d < bestDist {
			bestDist = d
			best = t.ID
		}
	}
	return best
}

func main() {
	// Google Maps API key can be supplied via environment variable.
	mapsKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if mapsKey == "" {
		log.Println("Warning: GOOGLE_MAPS_API_KEY is not set. Map features will be limited.")
	}

	// Default centre: Tokyo Station.
	store := newTaxiStore(35.6812, 139.7671)

	// Simulate taxi movement every 3 seconds.
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			store.move()
		}
	}()

	mux := http.NewServeMux()

	// Serve the static front-end.
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", fs)

	// GET /api/taxis – return current taxi positions.
	mux.HandleFunc("/api/taxis", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.all())
	})

	// POST /api/call – request the nearest taxi.
	mux.HandleFunc("/api/call", func(w http.ResponseWriter, r *http.Request) {
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

	// GET /api/maps-key – supply the Maps API key to the front-end at runtime.
	mux.HandleFunc("/api/maps-key", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"key": mapsKey})
	})

	addr := ":8080"
	log.Printf("Server listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
