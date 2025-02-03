package main

import (
    "encoding/json"
    "net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    switch r.Method {
    case http.MethodGet:
        json.NewEncoder(w).Encode(map[string]string{"data": "sample data"})
    case http.MethodPost:
        var input map[string]string
        json.NewDecoder(r.Body).Decode(&input)
        json.NewEncoder(w).Encode(map[string]interface{}{"received": input})
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func main() {
    http.HandleFunc("/api/health", healthHandler)
    http.HandleFunc("/api/data", dataHandler)
    http.ListenAndServe(":8080", nil)
}