CREATE TABLE parking_lots (
    id SERIAL PRIMARY KEY,
    total_spaces INT
);

CREATE TABLE parked_vehicles (
    id SERIAL PRIMARY KEY,
    parking_lot_id INT,
    slot INT,
    license_plate VARCHAR(20),
    entry_time TIMESTAMP,
    FOREIGN KEY (parking_lot_id) REFERENCES parking_lots(id)
);

CREATE TABLE parking_spaces (
    id SERIAL PRIMARY KEY,
    lot_id INT NOT NULL,
    number INT NOT NULL,
    occupied BOOLEAN DEFAULT false,
    in_maintenance BOOLEAN DEFAULT false,
    entry_time TIMESTAMP,
    FOREIGN KEY (lot_id) REFERENCES parking_lots(id)
);

CREATE INDEX idx_parking_spaces_lot_id_number ON parking_spaces (lot_id, number);



CREATE TABLE parking_transactions (
    id SERIAL PRIMARY KEY,
    lot_id INT NOT NULL,
    vehicle_license_plate VARCHAR(20) NOT NULL,
    entry_time TIMESTAMP NOT NULL,
    exit_time TIMESTAMP,
    fee INTEGER,
    CONSTRAINT fk_parking_transactions_lot_id FOREIGN KEY (lot_id) REFERENCES parking_lots(id)
);


--psql -U postgres -d db_vehicle_parking -h localhost -f migrations/migration.sql
