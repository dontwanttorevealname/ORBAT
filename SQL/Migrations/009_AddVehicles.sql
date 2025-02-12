-- +goose Up
CREATE TABLE vehicles (
    vehicle_id INTEGER PRIMARY KEY,
    vehicle_name TEXT,
    vehicle_type TEXT,
    vehicle_armament TEXT DEFAULT 'None',
    image_url TEXT
);

CREATE TABLE group_vehicles (
    group_id INTEGER,
    vehicle_id INTEGER,
    PRIMARY KEY (group_id, vehicle_id),
    FOREIGN KEY (group_id) REFERENCES groups(group_id),
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(vehicle_id)
);

CREATE TABLE vehicle_members (
    vehicle_id INTEGER,
    member_id INTEGER,
    PRIMARY KEY (vehicle_id, member_id),
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(vehicle_id),
    FOREIGN KEY (member_id) REFERENCES members(member_id)
);

-- +goose Down
DROP TABLE IF EXISTS vehicle_members;
DROP TABLE IF EXISTS group_vehicles;
DROP TABLE IF EXISTS vehicles;