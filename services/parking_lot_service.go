// services/parking_lot_service.go

package services

import (

	"parking_lot/storage"
)

type ParkingLotService struct {
	storage *storage.ParkingLotStorage
}

func NewParkingLotService(storage *storage.ParkingLotStorage) *ParkingLotService {
	return &ParkingLotService{storage: storage}
}

func (s *ParkingLotService) CreateParkingLot(totalSpaces int) (*storage.ParkingLot, error) {
	return s.storage.CreateParkingLot(totalSpaces)
}

func (s *ParkingLotService) ParkVehicle(parkingLotID int,LicensePlate string) (int, error) {
	return s.storage.ParkVehicle(parkingLotID,LicensePlate)
}

func (s *ParkingLotService) UnparkVehicle(parkingLotID int, LicensePlate string) (int, error) {
	return s.storage.UnparkVehicle(parkingLotID, LicensePlate)
}

func (s *ParkingLotService) ViewParkingLotStatus(parkingLotID int) (*storage.ParkingLotStatus, error) {
	return s.storage.ViewParkingLotStatus(parkingLotID)
}

func (s *ParkingLotService) ToggleMaintenance(parkingLotID, slotNumber int, inMaintenance bool) error {
	return s.storage.ToggleMaintenance(parkingLotID, slotNumber, inMaintenance)
}

func (s *ParkingLotService) GetReports(parkingLotID int) ([]*storage.DailyStats, error) {
	return s.storage.GetReports(parkingLotID)
}
