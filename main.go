package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
	"github.com/roylee0704/gron"
)

var db *sql.DB

type TaxiLocation struct {
	TaxiID    string  `json:"taxi_id"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type Place struct {
	PlaceID   int             `json:"place_id"`
	PlaceName string          `json:"place_name"`
	Polygon   json.RawMessage `json:"polygon"` // GeoJSON
}

func main() {
	var err error

	router := mux.NewRouter()

	connStr := "user=root dbname=subagiya1 password=secret host=localhost port=5431 sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize tables if they do not exist
	initTables()

	// CRUD Endpoints for Taxi Locations
	router.HandleFunc("/taxi", createTaxiLocation).Methods("POST")
	router.HandleFunc("/taxi", getAllTaxiLocations).Methods("GET")
	router.HandleFunc("/taxi/{id}", getTaxiLocation).Methods("GET")
	router.HandleFunc("/taxi/{id}", updateTaxiLocationCRUD).Methods("PUT")
	router.HandleFunc("/taxi/{id}", deleteTaxiLocation).Methods("DELETE")

	// CRUD Endpoints for Places
	router.HandleFunc("/place", createPlace).Methods("POST")
	router.HandleFunc("/places", getAllPlaces).Methods("GET")
	router.HandleFunc("/place/{id}", getPlace).Methods("GET")
	router.HandleFunc("/place/{id}", updatePlace).Methods("PUT")
	router.HandleFunc("/place/{id}", deletePlace).Methods("DELETE")

	// Existing Endpoints
	router.HandleFunc("/updateLocation", updateTaxiLocation).Methods("POST")
	router.HandleFunc("/getMapping", getMapping).Methods("GET")
	router.HandleFunc("/triggerMapping", triggerMapping).Methods("GET") // Added for manual trigger

	// Schedule mapping every 5 minutes
	c := gron.New()
	c.AddFunc(gron.Every(5*time.Minute), mapTaxiLocations)
	c.Start()

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// Initialize tables
func initTables() {
	tableCreationQueries := []string{
		`CREATE TABLE IF NOT EXISTS taxi_location (
            taxi_id VARCHAR PRIMARY KEY,
            longitude DOUBLE PRECISION,
            latitude DOUBLE PRECISION,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,
		`CREATE TABLE IF NOT EXISTS places (
            place_id SERIAL PRIMARY KEY,
            place_name VARCHAR,
            polygon JSONB
        )`,
		`CREATE TABLE IF NOT EXISTS mapping (
            map_id SERIAL PRIMARY KEY,
            taxi_id VARCHAR,
            place_id INTEGER,
            timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY(taxi_id) REFERENCES taxi_location(taxi_id),
            FOREIGN KEY(place_id) REFERENCES places(place_id)
        )`,
		`CREATE TABLE IF NOT EXISTS counters (
            taxi_id VARCHAR,
            place_id INTEGER,
            counter INTEGER DEFAULT 0,
            last_counted TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY(taxi_id, place_id),
            FOREIGN KEY(taxi_id) REFERENCES taxi_location(taxi_id),
            FOREIGN KEY(place_id) REFERENCES places(place_id)
        )`,
	}

	for _, query := range tableCreationQueries {
		if _, err := db.Exec(query); err != nil {
			log.Fatal("Failed to create table:", err)
		}
	}
}

//////////////////////
// CRUD for Taxis
//////////////////////

// Create Taxi Location
func createTaxiLocation(w http.ResponseWriter, r *http.Request) {
	var location TaxiLocation
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`INSERT INTO taxi_location (taxi_id, longitude, latitude, updated_at) 
        VALUES ($1, $2, $3, CURRENT_TIMESTAMP) 
        ON CONFLICT (taxi_id) DO NOTHING`,
		location.TaxiID, location.Longitude, location.Latitude)
	if err != nil {
		http.Error(w, "Failed to create taxi location", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Taxi location created.")
}

// Get All Taxi Locations
func getAllTaxiLocations(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT taxi_id, longitude, latitude FROM taxi_location")
	if err != nil {
		http.Error(w, "Failed to query taxi locations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var taxis []TaxiLocation
	for rows.Next() {
		var taxi TaxiLocation
		if err := rows.Scan(&taxi.TaxiID, &taxi.Longitude, &taxi.Latitude); err != nil {
			http.Error(w, "Failed to scan taxi location", http.StatusInternalServerError)
			return
		}
		taxis = append(taxis, taxi)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(taxis)
}

// Get Single Taxi Location
func getTaxiLocation(w http.ResponseWriter, r *http.Request) {
	taxiID := r.URL.Path[len("/taxi/"):]

	var taxi TaxiLocation
	err := db.QueryRow("SELECT taxi_id, longitude, latitude FROM taxi_location WHERE taxi_id = $1", taxiID).Scan(&taxi.TaxiID, &taxi.Longitude, &taxi.Latitude)
	if err == sql.ErrNoRows {
		http.Error(w, "Taxi not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to query taxi location", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(taxi)
}

// Update Taxi Location (CRUD)
func updateTaxiLocationCRUD(w http.ResponseWriter, r *http.Request) {
	taxiID := r.URL.Path[len("/taxi/"):]

	var location TaxiLocation
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`UPDATE taxi_location SET longitude = $1, latitude = $2, updated_at = CURRENT_TIMESTAMP 
        WHERE taxi_id = $3`,
		location.Longitude, location.Latitude, taxiID)
	if err != nil {
		http.Error(w, "Failed to update taxi location", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Taxi location updated.")
}

// Delete Taxi Location
func deleteTaxiLocation(w http.ResponseWriter, r *http.Request) {
	taxiID := r.URL.Path[len("/taxi/"):]

	_, err := db.Exec("DELETE FROM taxi_location WHERE taxi_id = $1", taxiID)
	if err != nil {
		http.Error(w, "Failed to delete taxi location", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Taxi location deleted.")
}

//////////////////////
// CRUD for Places
//////////////////////

// Create Place
func createPlace(w http.ResponseWriter, r *http.Request) {
	var place Place
	if err := json.NewDecoder(r.Body).Decode(&place); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var placeID int
	err := db.QueryRow(`INSERT INTO places (place_name, polygon) 
        VALUES ($1, $2) RETURNING place_id`,
		place.PlaceName, place.Polygon).Scan(&placeID)
	if err != nil {
		http.Error(w, "Failed to create place", http.StatusInternalServerError)
		return
	}

	place.PlaceID = placeID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(place)
}

// Get All Places
func getAllPlaces(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT place_id, place_name, polygon FROM places")
	if err != nil {
		http.Error(w, "Failed to query places", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var places []Place
	for rows.Next() {
		var place Place
		if err := rows.Scan(&place.PlaceID, &place.PlaceName, &place.Polygon); err != nil {
			http.Error(w, "Failed to scan place", http.StatusInternalServerError)
			return
		}
		places = append(places, place)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(places)
}

// Get Single Place
func getPlace(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/place/"):]
	placeID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid place ID", http.StatusBadRequest)
		return
	}

	var place Place
	err = db.QueryRow("SELECT place_id, place_name, polygon FROM places WHERE place_id = $1", placeID).
		Scan(&place.PlaceID, &place.PlaceName, &place.Polygon)
	if err == sql.ErrNoRows {
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to query place", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(place)
}

// Update Place
func updatePlace(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/place/"):]
	placeID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid place ID", http.StatusBadRequest)
		return
	}

	var place Place
	if err := json.NewDecoder(r.Body).Decode(&place); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`UPDATE places SET place_name = $1, polygon = $2 WHERE place_id = $3`,
		place.PlaceName, place.Polygon, placeID)
	if err != nil {
		http.Error(w, "Failed to update place", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Place updated.")
}

// Delete Place
func deletePlace(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/place/"):]
	placeID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid place ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM places WHERE place_id = $1", placeID)
	if err != nil {
		http.Error(w, "Failed to delete place", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Place deleted.")
}

//////////////////////
// Existing Functions
//////////////////////

// Endpoint to update taxi location
func updateTaxiLocation(w http.ResponseWriter, r *http.Request) {
	var location TaxiLocation
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`INSERT INTO taxi_location (taxi_id, longitude, latitude, updated_at) 
        VALUES ($1, $2, $3, CURRENT_TIMESTAMP) 
        ON CONFLICT (taxi_id) DO UPDATE 
        SET longitude = EXCLUDED.longitude, latitude = EXCLUDED.latitude, updated_at = CURRENT_TIMESTAMP`,
		location.TaxiID, location.Longitude, location.Latitude)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Taxi location updated.")
}

// Mapping function to assign taxi to place
func mapTaxiLocations() {
	rows, err := db.Query("SELECT taxi_id, longitude, latitude FROM taxi_location")
	if err != nil {
		log.Println("Error querying taxi locations:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var taxiID string
		var longitude, latitude float64
		if err := rows.Scan(&taxiID, &longitude, &latitude); err != nil {
			log.Println("Error scanning taxi location:", err)
			continue
		}

		log.Printf("Processing Taxi ID %s at (%f, %f)\n", taxiID, longitude, latitude)
		placeID, err := findPlace(longitude, latitude)
		if err != nil {
			log.Printf("No matching place found for Taxi ID %s at (%f, %f): %v\n", taxiID, longitude, latitude, err)
			continue
		}
		log.Printf("Mapping Taxi ID %s to Place ID %d\n", taxiID, placeID)
		updateMappingAndCounter(taxiID, placeID)
	}

	if err = rows.Err(); err != nil {
		log.Println("Row iteration error:", err)
	}
}

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   GeoJSONGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type GeoJSONGeometry struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

func findPlace(longitude, latitude float64) (int, error) {
	rows, err := db.Query("SELECT place_id, polygon FROM places")
	if err != nil {
		log.Println("Error querying places:", err)
		return 0, err
	}
	defer rows.Close()

	point := orb.Point{longitude, latitude}
	log.Printf("Checking point (%f, %f)\n", longitude, latitude) // Debug: show point being checked

	for rows.Next() {
		var placeID int
		var polygonData []byte
		if err := rows.Scan(&placeID, &polygonData); err != nil {
			log.Println("Error scanning place data:", err)
			return 0, err
		}

		var feature GeoJSONFeature
		if err := json.Unmarshal(polygonData, &feature); err != nil {
			log.Println("Error unmarshalling GeoJSON:", err)
			return 0, err
		}

		if feature.Geometry.Type != "Polygon" {
			continue
		}

		// Convert coordinates to orb.Ring
		var ring orb.Ring
		for _, coord := range feature.Geometry.Coordinates[0] {
			if len(coord) >= 2 {
				ring = append(ring, orb.Point{coord[0], coord[1]})
			}
		}

		polygon := orb.Polygon{ring}
		if planar.PolygonContains(polygon, point) {
			log.Printf("Match found: Point (%f, %f) is within Place ID %d\n", longitude, latitude, placeID)
			return placeID, nil
		} else {
			log.Printf("Point (%f, %f) is NOT within Place ID %d\n", longitude, latitude, placeID)
		}
	}

	log.Println("No matching place found")
	return 0, fmt.Errorf("no matching place found")
}

// Update mapping and increment counter
func updateMappingAndCounter(taxiID string, placeID int) {
	_, err := db.Exec("INSERT INTO mapping (taxi_id, place_id) VALUES ($1, $2)", taxiID, placeID)
	if err != nil {
		log.Println("Mapping insertion failed:", err)
	}

	var count int
	err = db.QueryRow("SELECT counter FROM counters WHERE taxi_id = $1 AND place_id = $2", taxiID, placeID).Scan(&count)
	if err == sql.ErrNoRows {
		_, err = db.Exec("INSERT INTO counters (taxi_id, place_id, counter, last_counted) VALUES ($1, $2, 1, CURRENT_TIMESTAMP)", taxiID, placeID)
	} else {
		_, err = db.Exec("UPDATE counters SET counter = counter + 1, last_counted = CURRENT_TIMESTAMP WHERE taxi_id = $1 AND place_id = $2", taxiID, placeID)
	}
	if err != nil {
		log.Println("Counter update failed:", err)
	}
}

// Endpoint to get current mappings
// Endpoint to manually trigger mapping
func triggerMapping(w http.ResponseWriter, r *http.Request) {
	mapTaxiLocations()
	fmt.Fprintf(w, "Mapping process triggered manually")
}

func getMapping(w http.ResponseWriter, r *http.Request) {
	query := `
        SELECT m.taxi_id, p.place_name, c.counter 
        FROM mapping m 
        JOIN places p ON m.place_id = p.place_id 
        JOIN counters c ON m.taxi_id = c.taxi_id AND m.place_id = c.place_id
    `
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Failed to query mappings: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var mappings []map[string]interface{}
	count := 0
	for rows.Next() {
		var taxiID, placeName string
		var counter int
		if err := rows.Scan(&taxiID, &placeName, &counter); err != nil {
			http.Error(w, "Failed to scan mapping: "+err.Error(), http.StatusInternalServerError)
			return
		}
		mappings = append(mappings, map[string]interface{}{
			"taxi_id": taxiID,
			"place":   placeName,
			"counter": counter,
		})
		count++
	}

	log.Printf("Fetched %d mappings\n", count)

	if err = rows.Err(); err != nil {
		http.Error(w, "Row iteration error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// If no mappings found, return an empty array instead of null
	w.Header().Set("Content-Type", "application/json")
	if count == 0 {
		json.NewEncoder(w).Encode([]map[string]interface{}{})
		return
	}
	json.NewEncoder(w).Encode(mappings)
}
