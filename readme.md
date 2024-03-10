curl -X POST -H "Content-Type: application/json" -d '{"totalSpaces": 10}' http://localhost:8081/createParkingLot

curl -X POST -H "Content-Type: application/json" -d '{"parkingLotID": 6, "licensePlate": "ABC123"}' http://localhost:8081/parkVehicle

curl -X POST -H "Content-Type: application/json" -d '{"parkingLotID": 6, "licensePlate": "ABC123"}' http://localhost:8081/unparkVehicle

curl -X GET -H "Content-Type: application/json" -d '{"parkingLotID": 1}' http://localhost:8081/viewParkingLotStatus

curl -X POST -H "Content-Type: application/json" -d '{"parkingLotID": 6, "slotNumber": 2,"inMaintenance":true}' http://localhost:8081/toggleMaintenance

curl -X GET -H "Content-Type: application/json" -d '{"parkingLotID": 6}' http://localhost:8081/getTotalStats
