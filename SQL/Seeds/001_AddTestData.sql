-- +goose Up

INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber, image_url) VALUES
(1000, 'Test Rifle', 'Test Type', 'Test Caliber', 'test-rifle.jpg'),
(1001, 'Test Machine Gun', 'Test Type', 'Test Caliber', 'test-mg.jpg');

INSERT INTO vehicles (vehicle_id, vehicle_name, vehicle_type, vehicle_armament, image_url) VALUES
(1000, 'Test Vehicle 1', 'Test Type', 'Test Gun', 'test-vehicle-1.jpg'),
(1001, 'Test Vehicle 2', 'Test Type', 'None', 'test-vehicle-2.jpg');

INSERT INTO members (member_id, member_role, member_rank) VALUES
(1000, 'Test Leader', 'Test Rank'),
(1001, 'Test Gunner', 'Test Rank'),
(1002, 'Test Driver', 'Test Rank');

INSERT INTO members_weapons (member_id, weapon_id) VALUES
(1000, 1000),
(1001, 1001);

INSERT INTO teams (team_id, team_name, team_size) VALUES
(1000, 'Test Team', 2);

INSERT INTO team_members (team_id, member_id) VALUES
(1000, 1001),
(1000, 1002);

INSERT INTO groups (group_id, group_name, group_size, group_nationality) VALUES
(1000, 'Test Group', 3, 'Test Nation');

INSERT INTO group_members (group_id, member_id, team_id) VALUES
(1000, 1000, NULL),    -- Test Leader
(1000, NULL, 1000);    -- Test Team

INSERT INTO group_vehicles (instance_id, group_id, vehicle_id) VALUES
(1000, 1000, 1000),
(1001, 1000, 1001);

INSERT INTO vehicle_members (instance_id, member_id) VALUES
(1000, 1001),  -- Gunner in first vehicle
(1000, 1002);  -- Driver in first vehicle


-- +goose Down
DELETE FROM vehicle_members WHERE instance_id IN (1000, 1001);
DELETE FROM group_vehicles WHERE instance_id IN (1000, 1001);
DELETE FROM group_members WHERE group_id = 1000;
DELETE FROM groups WHERE group_id = 1000;
DELETE FROM team_members WHERE team_id = 1000;
DELETE FROM teams WHERE team_id = 1000;
DELETE FROM members_weapons WHERE member_id IN (1000, 1001, 1002);
DELETE FROM members WHERE member_id IN (1000, 1001, 1002);
DELETE FROM vehicles WHERE vehicle_id IN (1000, 1001);
DELETE FROM weapons WHERE weapon_id IN (1000, 1001); 
