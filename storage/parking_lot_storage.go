package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

const (
	dbUser            = "postgres"
	dbPassword        = "password"
	dbName            = "db_vehicle_parking"
	ParkingFeeperHour = 10
)

// ParkingLot represents a parking lot with parking spaces.
type ParkingLot struct {
	ID          int
	TotalSpaces int
	Spaces      []ParkingSpace
}

// ParkingSpace represents a parking space in a parking lot.
type ParkingSpace struct {
	Number        int
	InMaintenance bool
	Occupied      bool
	EntryTime     time.Time
}

// ParkingLotStatus represents the current status of a parking lot.
type ParkingLotStatus struct {
	ParkedVehicles map[int]VehicleStatus
}

// VehicleStatus represents the status of a parked vehicle.
type VehicleStatus struct {
	Vehicle    string
	SlotNumber int
	EntryTime  time.Time
}

// DailyStats represents the total statistics for a parking lot per day.
type DailyStats struct {
	Day              time.Time `json:"day"`
	TotalVehicles    int       `json:"total_vehicles"`
	TotalParkingTime float64   `json:"total_parking_time"`
	TotalFee         int       `json:"total_fee"`
}

// Vehicle represents a parked vehicle.
type Vehicle struct {
	ID           int
	ParkingLotID int
	SlotNumber   int
	EntryTime    time.Time
}

// ParkingLotStorage provides storage for parking lots and vehicles.
type ParkingLotStorage struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewParkingLotStorage creates a new instance of ParkingLotStorage.
func NewParkingLotStorage() (*ParkingLotStorage, error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &ParkingLotStorage{db: db}, nil
}

// CreateParkingLot creates a new parking lot with the specified total spaces.
func (s *ParkingLotStorage) CreateParkingLot(totalSpaces int) (*ParkingLot, error) {
	var parkingLotID int

	err := s.db.QueryRow("INSERT INTO parking_lots(total_spaces) VALUES($1) RETURNING id", totalSpaces).Scan(&parkingLotID)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var parkingSpaces []ParkingSpace
	for i := 1; i <= totalSpaces; i++ {
		_, err := s.db.Exec(`
			INSERT INTO parking_spaces(lot_id, number)
			VALUES($1, $2)
		`, parkingLotID, i)

		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		parkingSpaces = append(parkingSpaces, ParkingSpace{
			Number: i,
		})
	}

	parkingLot := &ParkingLot{
		ID:          parkingLotID,
		TotalSpaces: totalSpaces,
		Spaces:      parkingSpaces,
	}

	return parkingLot, nil
}

// ParkVehicle parks a vehicle in the nearest available slot in the specified parking lot.
func (s *ParkingLotStorage) ParkVehicle(parkingLotID int, LicensePlate string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var totalSpaces int
	err := s.db.QueryRow("SELECT total_spaces FROM parking_lots WHERE id = $1", parkingLotID).Scan(&totalSpaces)
	if err != nil {
		return 0, errors.New("parking lot not found")
	}

	var nearestSoltID int
	err = s.db.QueryRow("SELECT id from parking_spaces WHERE lot_id = $1 AND NOT occupied AND NOT in_maintenance ORDER BY number LIMIT 1", parkingLotID).Scan(&nearestSoltID)
	if err != nil {
		return 0, errors.New("nearest available slot not found")
	}

	var slotNumber int
	err = s.db.QueryRow(`
		UPDATE parking_spaces
		SET occupied = true, entry_time = NOW()
		WHERE id = $1
		RETURNING number
	`, nearestSoltID).Scan(&slotNumber)

	if err != nil {
		return 0, errors.New("failed to occupy parking space")
	}
	var vehicleId int
	err = s.db.QueryRow("INSERT INTO parked_vehicles(parking_lot_id,slot,license_plate,entry_time) VALUES($1,$2,$3,NOW()) RETURNING id", parkingLotID, slotNumber, LicensePlate).Scan(&vehicleId)
	if err != nil {
		log.Fatal(err)
		return 0, err
	}

	return slotNumber, nil
}

// UnparkVehicle unparks a vehicle from the specified parking lot.
// It returns the parking fee calculated based on the entry time.
func (s *ParkingLotStorage) UnparkVehicle(parkingLotID int, LicensePlate string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var parkingSpaceID int
	err := s.db.QueryRow("SELECT parking_spaces.id FROM parked_vehicles LEFT JOIN parking_spaces ON parking_spaces.lot_id=parked_vehicles.parking_lot_id and parked_vehicles.slot=parking_spaces.number WHERE parking_spaces.lot_id = $1 AND parked_vehicles.license_plate=$2 AND occupied=TRUE", parkingLotID, LicensePlate).Scan(&parkingSpaceID)
	if err != nil {
		return 0, errors.New("required parked vehicle lot not found")
	}

	var entryTime time.Time
	err = s.db.QueryRow(`
		UPDATE parking_spaces
		SET occupied = false
		WHERE id = $1 
		RETURNING entry_time
	`, parkingSpaceID).Scan(&entryTime)

	if err != nil {
		return 0, errors.New("failed to unpark vehicle")
	}

	// Calculate the parking fee and update the parking transaction
	exitTime := time.Now().In(entryTime.Location())
	parkingTime := exitTime.Sub(entryTime)
	fee := int(math.Ceil(parkingTime.Hours())) * 10

	log.Println("*******************************")
	log.Println(exitTime, entryTime, int(math.Ceil(parkingTime.Hours())))

	_, err = s.db.Exec(`
		INSERT INTO parking_transactions (lot_id, vehicle_license_plate,fee, entry_time,exit_time)
		VALUES ($1, $2, $3, $4, NOW())
	`, parkingLotID, LicensePlate, fee, entryTime)

	if err != nil {
		log.Fatal(err)
		return 0, err
	}

	return fee, nil
}

// ViewParkingLotStatus retrieves the current status of the specified parking lot.
func (s *ParkingLotStorage) ViewParkingLotStatus(parkingLotID int) (*ParkingLotStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalSpaces int
	err := s.db.QueryRow("SELECT total_spaces FROM parking_lots WHERE id = $1", parkingLotID).Scan(&totalSpaces)
	if err != nil {
		return nil, errors.New("parking lot not found")
	}

	rows, err := s.db.Query(`
		SELECT number, occupied, parking_spaces.entry_time,license_plate
		FROM parking_spaces
		LEFT JOIN parked_vehicles ON parking_spaces.lot_id=parked_vehicles.parking_lot_id and parked_vehicles.slot=parking_spaces.number
		WHERE lot_id = $1 and occupied =TRUE
	`, parkingLotID)

	if err != nil {
		return nil, errors.New("failed to retrieve parking lot status")
	}
	defer rows.Close()

	status := &ParkingLotStatus{
		ParkedVehicles: make(map[int]VehicleStatus),
	}
	index := 0
	for rows.Next() {
		index++
		var vehicle string
		var spaceNumber int
		var occupied bool
		var entryTime time.Time

		err := rows.Scan(&spaceNumber, &occupied, &entryTime, &vehicle)
		if err != nil {
			log.Println(err)
			return nil, errors.New("failed to  parking lot status")
		}

		if occupied {
			status.ParkedVehicles[index] = VehicleStatus{
				Vehicle:    vehicle,
				SlotNumber: spaceNumber,
				EntryTime:  entryTime,
			}
		}
	}

	return status, nil
}

// ToggleMaintenance toggles the maintenance mode of a parking space in the specified parking lot.
func (s *ParkingLotStorage) ToggleMaintenance(parkingLotID, slotNumber int, inMaintenance bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var totalSpaces int
	err := s.db.QueryRow("SELECT total_spaces FROM parking_lots WHERE id = $1", parkingLotID).Scan(&totalSpaces)
	if err != nil {
		return errors.New("parking lot not found")
	}
	log.Println(inMaintenance, parkingLotID, slotNumber)
	_, err = s.db.Exec(`
		UPDATE parking_spaces
		SET in_maintenance = $1
		WHERE lot_id = $2 AND number = $3
	`, inMaintenance, parkingLotID, slotNumber)

	if err != nil {
		return errors.New("failed to toggle maintenance mode")
	}

	return nil
}

// GetReports retrieves total statistics for the specified parking lot.
func (s *ParkingLotStorage) GetReports(parkingLotID int) ([]*DailyStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalSpaces int
	err := s.db.QueryRow("SELECT total_spaces FROM parking_lots WHERE id = $1", parkingLotID).Scan(&totalSpaces)
	if err != nil {
		return nil, errors.New("parking lot not found")
	}

	rows, err := s.db.Query(`
		SELECT
			DATE(parking_transactions.exit_time) AS day,
			COUNT(*) AS total_vehicles,
			COALESCE(SUM(EXTRACT(EPOCH FROM (parking_transactions.exit_time - parking_transactions.entry_time)) / 3600), 0) AS total_parking_time,
			COALESCE(SUM(parking_transactions.fee), 0) AS total_fee
		FROM parking_transactions
		WHERE lot_id = $1
		GROUP BY day
		ORDER BY day
	`, parkingLotID)
	if err != nil {
		return nil, errors.New("failed to retrieve dawise total statistics")
	}
	defer rows.Close()

	var dailyStatsList []*DailyStats
	for rows.Next() {
		var dailyStats DailyStats
		if err := rows.Scan(&dailyStats.Day, &dailyStats.TotalVehicles, &dailyStats.TotalParkingTime, &dailyStats.TotalFee); err != nil {
			return nil, errors.New("failed to day wise total statitics")
		}
		dailyStatsList = append(dailyStatsList, &dailyStats)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("error processing daywise total statitics")
	}

	return dailyStatsList, nil
}
