package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"parking_lot/services"
	"parking_lot/storage"

	"github.com/gorilla/mux"
)

func main() {

	// Initialize storage n servicce
	parkingLotStorage, err := storage.NewParkingLotStorage()
	if err != nil {
		log.Fatal("Failed to initialize storage:", err)
	}
	parkingLotService := services.NewParkingLotService(parkingLotStorage)

	router := mux.NewRouter()

	// Endpoints
	router.HandleFunc("/createParkingLot", createParkingLotHandler(parkingLotService)).Methods("POST")

	router.HandleFunc("/parkVehicle", parkVehicleHandler(parkingLotService)).Methods("POST")

	router.HandleFunc("/unparkVehicle", unparkVehicleHandler(parkingLotService)).Methods("POST")

	router.HandleFunc("/viewParkingLotStatus", viewParkingLotStatusHandler(parkingLotService)).Methods("GET")

	router.HandleFunc("/toggleMaintenance", toggleMaintenanceHandler(parkingLotService)).Methods("POST")

	router.HandleFunc("/getTotalStats", getTotalStatsHandler(parkingLotService)).Methods("GET")

	fmt.Println("*************************************")
	fmt.Println("Server is running on :8081...")
	http.ListenAndServe(":8081", router)
}

// Handler for creating a parking lot
func createParkingLotHandler(service *services.ParkingLotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			TotalSpaces int `json:"totalSpaces"`
		}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		parkingLot, err := service.CreateParkingLot(request.TotalSpaces)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create parking lot: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(parkingLot)
	}
}

// For parking a vehicle
func parkVehicleHandler(service *services.ParkingLotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ParkingLotID int    `json:"parkingLotID"`
			LicensePlate string `json:"licensePlate"`
		}

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		slotNumber, err := service.ParkVehicle(request.ParkingLotID, request.LicensePlate)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to park vehicle: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			SlotNumber int `json:"slotNumber"`
		}{SlotNumber: slotNumber})
	}
}

// For unparking a vehicle
func unparkVehicleHandler(service *services.ParkingLotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ParkingLotID int    `json:"parkingLotID"`
			LicensePlate string `json:"licensePlate"`
		}

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		fee, err := service.UnparkVehicle(request.ParkingLotID, request.LicensePlate)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to unpark vehicle: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Fee int `json:"fee"`
		}{Fee: fee})
	}
}

// For viewing parking lot status
func viewParkingLotStatusHandler(service *services.ParkingLotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ParkingLotID int `json:"parkingLotID"`
		}

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		status, err := service.ViewParkingLotStatus(request.ParkingLotID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get parking lot status: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}

// For toggling maintenance mode
func toggleMaintenanceHandler(service *services.ParkingLotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ParkingLotID  int  `json:"parkingLotID"`
			SlotNumber    int  `json:"slotNumber"`
			InMaintenance bool `json:"inMaintenance"`
		}

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		err = service.ToggleMaintenance(request.ParkingLotID, request.SlotNumber, request.InMaintenance)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to toggle maintenance mode: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Message string `json:"message"`
		}{Message: "Maintenance mode toggled successfully"})
	}
}

// for getting total statistics
func getTotalStatsHandler(service *services.ParkingLotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ParkingLotID int `json:"parkingLotID"`
		}

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		stats, err := service.GetReports(request.ParkingLotID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get total statistics: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}
