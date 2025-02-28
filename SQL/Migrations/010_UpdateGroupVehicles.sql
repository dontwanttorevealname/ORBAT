-- +goose Up
-- Drop the existing table
DROP TABLE IF EXISTS group_vehicles;

-- Recreate with an instance_id column
CREATE TABLE group_vehicles (
    instance_id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER,
    vehicle_id INTEGER,
    FOREIGN KEY (group_id) REFERENCES groups(group_id),
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(vehicle_id)
);

-- Update vehicle_members table
DROP TABLE IF EXISTS vehicle_members;
CREATE TABLE vehicle_members (
    instance_id INTEGER,
    member_id INTEGER,
    PRIMARY KEY (instance_id, member_id),
    FOREIGN KEY (instance_id) REFERENCES group_vehicles(instance_id),
    FOREIGN KEY (member_id) REFERENCES members(member_id)
);

-- +goose Down
DROP TABLE IF EXISTS vehicle_members;
DROP TABLE IF EXISTS group_vehicles;