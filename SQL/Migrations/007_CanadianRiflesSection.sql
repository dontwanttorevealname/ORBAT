-- +goose Up
-- Add new weapons
INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber) VALUES
(10, 'C7A2', 'Assault Rifle', '5.56mm'),
(11, 'C9A2', 'Light Machine Gun', '5.56mm');

-- Members
INSERT INTO members (member_id, member_role, member_rank) VALUES
(58, 'Section Commander', 'Master Corporal'),
(59, '2IC', 'Corporal'),
-- Rifle Group
(60, 'Rifleman', 'Private'),
(61, 'Rifleman', 'Private'),
(62, 'Rifleman', 'Private'),
-- Gun Group
(63, 'Machine Gunner', 'Private'),
(64, 'Machine Gunner', 'Private');

-- Weapon assignments
INSERT INTO members_weapons (member_id, weapon_id) VALUES
(58, 7),  -- Section Commander - C7A2
(59, 7),  -- 2IC - C7A2
(60, 7),  -- Rifleman - C7A2
(61, 7),  -- Rifleman - C7A2
(62, 7),  -- Rifleman - C7A2
(63, 8),  -- Machine Gunner - C9A2
(64, 8);  -- Machine Gunner - C9A2

-- Teams
INSERT INTO teams (team_id, team_name, team_size) VALUES
(12, 'Rifle Group', 3),
(13, 'Gun Group', 2);

-- Team members
INSERT INTO team_members (team_id, member_id) VALUES
-- Rifle Group
(12, 60), (12, 61), (12, 62),
-- Gun Group
(13, 63), (13, 64);

-- Group
INSERT INTO groups (group_id, group_name, group_size, group_nationality) VALUES
(6, 'Canadian Infantry Section', 7, 'Canada');

-- Group members
INSERT INTO group_members (group_id, member_id, team_id) VALUES
(6, 58, NULL),     -- Section Commander
(6, 59, NULL),     -- 2IC
(6, NULL, 12),     -- Rifle Group
(6, NULL, 13);     -- Gun Group

-- +goose Down
DELETE FROM group_members WHERE group_id = 6;
DELETE FROM groups WHERE group_id = 6;
DELETE FROM team_members WHERE team_id IN (12, 13);
DELETE FROM teams WHERE team_id IN (12, 13);
DELETE FROM members_weapons WHERE member_id BETWEEN 58 AND 64;
DELETE FROM members WHERE member_id BETWEEN 58 AND 64;
DELETE FROM weapons WHERE weapon_id IN (7, 8);