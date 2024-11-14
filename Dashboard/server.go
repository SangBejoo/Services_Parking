package dashboard

import (
    "net/http"
    "github.com/gorilla/mux"
)


// getVehiclesHandler handles the /api/vehicles endpoint
func getVehiclesHandler(w http.ResponseWriter, r *http.Request) {
    // Implement your handler logic here
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Vehicles data"))
}

// dashboard/server.go
func setupDashboard() *http.Server {
    router := mux.NewRouter()
    
    // Serve static files
    router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", 
        http.FileServer(http.Dir("dashboard/static"))))
    
    // API endpoints
    router.HandleFunc("/api/vehicles", getVehiclesHandler).Methods("GET")
    
    return &http.Server{
        Addr:    ":8080",
        Handler: router,
    }
}
