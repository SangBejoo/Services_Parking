// package main

// import (
//     "encoding/json"
//     "fmt"
//     "log"
//     "math/rand"
//     "net/http"
//     "strconv" // Import tambahan
//     "sync"
//     "time"

//     "github.com/gorilla/websocket"
//     "github.com/jasonlvhit/gocron"
// )

// // Car represents a car with its location and timestamp
// type Car struct {
//     ID        int       `json:"id"`
//     Longitude float64   `json:"longitude"`
//     Latitude  float64   `json:"latitude"`
//     Timestamp time.Time `json:"timestamp"`
//     Source    string    `json:"source"` // "dummy" or "snapshot"
// }

// // generateDummyData generates dummy car data
// func generateDummyData() []Car {
//     baseTime := time.Now()
//     return []Car{
//         {ID: 1, Longitude: 100.0 + rand.Float64(), Latitude: 0.0 + rand.Float64(), Timestamp: baseTime, Source: "dummy"},
//         {ID: 2, Longitude: 101.0 + rand.Float64(), Latitude: 1.0 + rand.Float64(), Timestamp: baseTime, Source: "dummy"},
//         {ID: 3, Longitude: 102.0 + rand.Float64(), Latitude: 2.0 + rand.Float64(), Timestamp: baseTime, Source: "dummy"},
//         {ID: 4, Longitude: 103.0 + rand.Float64(), Latitude: 3.0 + rand.Float64(), Timestamp: baseTime, Source: "dummy"},
//         {ID: 5, Longitude: 104.0 + rand.Float64(), Latitude: 4.0 + rand.Float64(), Timestamp: baseTime, Source: "dummy"},
//     }
// }

// // SpatialHash structure
// type SpatialHash struct {
//     gridSize float64
//     buckets  map[string][]Car
//     carMap   map[int]Car            // Peta ID mobil ke objek Car
//     mu       sync.RWMutex
//     clients  map[*websocket.Conn]bool
// }

// var upgrader = websocket.Upgrader{
//     CheckOrigin: func(r *http.Request) bool {
//         return true
//     },
// }

// // NewSpatialHash creates a new spatial hash
// func NewSpatialHash(gridSize float64) *SpatialHash {
//     return &SpatialHash{
//         gridSize: gridSize,
//         buckets:  make(map[string][]Car),
//         carMap:   make(map[int]Car),
//         clients:  make(map[*websocket.Conn]bool),
//     }
// }

// // AddOrUpdateCar adds or updates a car in the spatial hash
// func (sh *SpatialHash) AddOrUpdateCar(car Car) {
//     sh.mu.Lock()
//     defer sh.mu.Unlock()

//     // Periksa apakah mobil sudah ada
//     existingCar, exists := sh.carMap[car.ID]
//     if exists {
//         // Hapus mobil dari bucket lama
//         oldKey := sh.hash(existingCar.Longitude, existingCar.Latitude)
//         sh.removeCarFromBucket(oldKey, car.ID)
//     }

//     // Tambahkan atau perbarui mobil di carMap
//     sh.carMap[car.ID] = car

//     // Tambahkan mobil ke bucket baru
//     key := sh.hash(car.Longitude, car.Latitude)
//     sh.buckets[key] = append(sh.buckets[key], car)
// }

// // removeCarFromBucket menghapus mobil dari bucket tertentu berdasarkan ID
// func (sh *SpatialHash) removeCarFromBucket(key string, carID int) {
//     cars := sh.buckets[key]
//     for i, car := range cars {
//         if car.ID == carID {
//             // Hapus mobil dari slice
//             sh.buckets[key] = append(cars[:i], cars[i+1:]...)
//             break
//         }
//     }
// }

// // hash generates a hash key for given longitude and latitude
// func (sh *SpatialHash) hash(longitude, latitude float64) string {
//     x := int(longitude / sh.gridSize)
//     y := int(latitude / sh.gridSize)
//     return fmt.Sprintf("%d:%d", x, y)
// }

// // handleWebSocket handles websocket connections
// func (sh *SpatialHash) handleWebSocket(w http.ResponseWriter, r *http.Request) {
//     conn, err := upgrader.Upgrade(w, r, nil)
//     if err != nil {
//         log.Printf("Websocket upgrade failed: %v", err)
//         return
//     }
//     defer conn.Close()

//     sh.mu.Lock()
//     sh.clients[conn] = true
//     sh.mu.Unlock()

//     // Send current cars to the client
//     err = conn.WriteJSON(sh.GetAllCars())
//     if err != nil {
//         log.Printf("Websocket write error: %v", err)
//         conn.Close()
//         sh.mu.Lock()
//         delete(sh.clients, conn)
//         sh.mu.Unlock()
//         return
//     }

//     // Remove client on disconnect
//     defer func() {
//         sh.mu.Lock()
//         delete(sh.clients, conn)
//         sh.mu.Unlock()
//     }()

//     // Keep connection alive
//     for {
//         _, _, err := conn.ReadMessage()
//         if err != nil {
//             break
//         }
//     }
// }

// // BroadcastUpdate sends updates to all connected clients
// func (sh *SpatialHash) BroadcastUpdate(cars []Car) {
//     sh.mu.RLock()
//     defer sh.mu.RUnlock()

//     for client := range sh.clients {
//         err := client.WriteJSON(cars)
//         if err != nil {
//             log.Printf("Websocket write error: %v", err)
//             client.Close()
//             delete(sh.clients, client)
//         }
//     }
// }

// // GetAllCars returns all cars in the spatial hash
// func (sh *SpatialHash) GetAllCars() []Car {
//     sh.mu.RLock()
//     defer sh.mu.RUnlock()

//     allCars := make([]Car, 0, len(sh.carMap))
//     for _, car := range sh.carMap {
//         allCars = append(allCars, car)
//     }
//     return allCars
// }

// // SimulateSnapshot simulates getting data from an API
// func SimulateSnapshot() []Car {
//     // Simulate some variation in the data
//     baseTime := time.Now()
//     return []Car{
//         {ID: 1, Longitude: 106.84513 + rand.Float64()*0.01, Latitude: -6.21462 + rand.Float64()*0.01, Timestamp: baseTime, Source: "snapshot"},
//         {ID: 2, Longitude: 106.84513 + rand.Float64()*0.01, Latitude: -6.21462 + rand.Float64()*0.01, Timestamp: baseTime, Source: "snapshot"},
//         {ID: 3, Longitude: 106.84513 + rand.Float64()*0.01, Latitude: -6.21462 + rand.Float64()*0.01, Timestamp: baseTime, Source: "snapshot"},
//         {ID: 4, Longitude: 106.84513 + rand.Float64()*0.01, Latitude: -6.21462 + rand.Float64()*0.01, Timestamp: baseTime, Source: "snapshot"},
//         {ID: 5, Longitude: 106.84513 + rand.Float64()*0.01, Latitude: -6.21462 + rand.Float64()*0.01, Timestamp: baseTime, Source: "snapshot"},
//         // Tambahkan mobil baru untuk pengujian
//         {ID: 6, Longitude: 107.00000, Latitude: -6.30000, Timestamp: baseTime, Source: "snapshot"},
//     }
// }

// // UpdateData updates both dummy and snapshot data by comparing existing cars
// func (sh *SpatialHash) UpdateData() {
//     // Mendapatkan data snapshot baru
//     snapshotData := SimulateSnapshot()

//     sh.mu.Lock()
//     defer sh.mu.Unlock()

//     for _, newCar := range snapshotData {
//         existingCar, exists := sh.carMap[newCar.ID]
//         if exists {
//             // Bandingkan lokasi atau timestamp jika diperlukan
//             if existingCar.Longitude != newCar.Longitude || existingCar.Latitude != newCar.Latitude {
//                 // Perbarui lokasi dan timestamp
//                 sh.carMap[newCar.ID] = newCar

//                 // Hapus dari bucket lama
//                 oldKey := sh.hash(existingCar.Longitude, existingCar.Latitude)
//                 sh.removeCarFromBucket(oldKey, newCar.ID)

//                 // Tambahkan ke bucket baru
//                 newKey := sh.hash(newCar.Longitude, newCar.Latitude)
//                 sh.buckets[newKey] = append(sh.buckets[newKey], newCar)
//             }
//             // Jika tidak ada perubahan, lewati
//         } else {
//             // Tambahkan mobil baru
//             sh.carMap[newCar.ID] = newCar
//             key := sh.hash(newCar.Longitude, newCar.Latitude)
//             sh.buckets[key] = append(sh.buckets[key], newCar)
//         }
//     }

//     // Opsional: Menghapus mobil yang tidak ada di snapshot baru
//     // Anda dapat menambahkan logika di sini jika diperlukan

//     // Broadcast update ke semua klien
//     sh.BroadcastUpdate(sh.GetAllCars())
// }

// // HTML template for the dashboard
// const dashboardHTML = `
// <!DOCTYPE html>
// <html>
// <head>
//     <title>Dashboard</title>
//     <script>
//         var ws = new WebSocket("ws://" + window.location.host + "/ws");
//         ws.onmessage = function(event) {
//             var data = JSON.parse(event.data);
//             console.log(data);
//             // Update your dashboard dengan data baru
//         };
//     </script>
// </head>
// <body>
//     <h1>Dashboard</h1>
//     <p>Data akan dicetak di konsol browser.</p>
// </body>
// </html>
// `

// // addCarHandler handles adding a new car via POST request
// func (sh *SpatialHash) addCarHandler(w http.ResponseWriter, r *http.Request) {
//     var car Car
//     if err := json.NewDecoder(r.Body).Decode(&car); err != nil {
//         http.Error(w, err.Error(), http.StatusBadRequest)
//         return
//     }
//     sh.AddOrUpdateCar(car)
//     w.WriteHeader(http.StatusCreated)
// }

// // getCarByIDHandler handles retrieving a single car by ID
// func getCarByIDHandler(w http.ResponseWriter, r *http.Request, idStr string, sh *SpatialHash) {
//     // Konversi ID dari string ke integer
//     id, err := strconv.Atoi(idStr)
//     if err != nil {
//         http.Error(w, "Invalid car ID", http.StatusBadRequest)
//         return
//     }

//     sh.mu.RLock()
//     car, exists := sh.carMap[id]
//     sh.mu.RUnlock()

//     if !exists {
//         http.Error(w, "Car not found", http.StatusNotFound)
//         return
//     }

//     w.Header().Set("Content-Type", "application/json")
//     json.NewEncoder(w).Encode(car)
// }

// func main() {
//     sh := NewSpatialHash(1.0)

//     // Initialize dengan data dummy
//     cars := generateDummyData()
//     for _, car := range cars {
//         sh.AddOrUpdateCar(car)
//     }

//     // Set up scheduling
//     scheduler := gocron.NewScheduler()
//     scheduler.Every(5).Minutes().Do(func() {
//         log.Println("Updating dashboard data...")
//         sh.UpdateData()
//     })
//     go scheduler.Start()

//     // HTTP handlers
//     http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//         fmt.Fprint(w, dashboardHTML)
//     })
//     http.HandleFunc("/ws", sh.handleWebSocket)
//     http.HandleFunc("/cars", func(w http.ResponseWriter, r *http.Request) {
//         switch r.Method {
//         case http.MethodGet:
//             // Cek apakah ada parameter ID di query
//             idStr := r.URL.Query().Get("id")
//             if idStr != "" {
//                 getCarByIDHandler(w, r, idStr, sh)
//                 return
//             }
//             // Jika tidak ada parameter ID, return semua mobil
//             w.Header().Set("Content-Type", "application/json")
//             json.NewEncoder(w).Encode(sh.GetAllCars())
//         case http.MethodPost:
//             sh.addCarHandler(w, r)
//         default:
//             http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//         }
//     })

//     // Start server
//     log.Println("Server starting on :8080")
//     log.Fatal(http.ListenAndServe(":8080", nil))
// }