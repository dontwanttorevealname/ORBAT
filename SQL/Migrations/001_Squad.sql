-- +goose Up
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS members;
DROP TABLE IF EXISTS weapons;



CREATE TABLE weapons (
    weapon_id INTEGER PRIMARY KEY,
    weapon_name TEXT,
    weapon_type TEXT,
    weapon_caliber TEXT
);

CREATE TABLE members (
    member_id INTEGER PRIMARY KEY,
    member_role TEXT,
    member_rank TEXT,
    weapon_id INTEGER,
    FOREIGN KEY (weapon_id) REFERENCES weapons(weapon_id)
);

CREATE TABLE teams (
    team_id INTEGER PRIMARY KEY,
    team_name TEXT,
    team_size int
);

CREATE TABLE team_members (
    team_id INTEGER,
    member_id INTEGER,
    PRIMARY KEY (team_id, member_id),
    FOREIGN KEY (team_id) REFERENCES teams(team_id),
    FOREIGN KEY (member_id) REFERENCES members(member_id)
);


CREATE TABLE groups (
    group_id INTEGER PRIMARY KEY,
    group_name TEXT,
    group_size int,
    group_nationality TEXT
);

CREATE TABLE group_members (
    group_id INTEGER,
    member_id INTEGER NULL,
    team_id INTEGER NULL,
    PRIMARY KEY (group_id, member_id, team_id),
    FOREIGN KEY (group_id) REFERENCES groups(group_id),
    FOREIGN KEY (member_id) REFERENCES members(member_id),
    FOREIGN KEY (team_id) REFERENCES teams(team_id),
    CHECK ((member_id IS NULL AND team_id IS NOT NULL) OR 
           (member_id IS NOT NULL AND team_id IS NULL))
);


-- Test data for weapons
INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber) VALUES
(1, 'M4A1', 'Assault Rifle', '5.56mm'),
(2, 'M249', 'Squad Automatic Weapon', '5.56mm'),
(3, 'M320', 'Grenade Launcher', '40mm'),
(4, 'M110', 'Designated Marksman Rifle', '7.62mm');

-- Test data for members (US Army Ranger Rifle Squad)
INSERT INTO members (member_id, member_role, member_rank, weapon_id) VALUES
(1, 'Squad Leader', 'Staff Sergeant', 1),
-- Alpha Team
(2, 'Team Leader', 'Sergeant', 1),
(3, 'Automatic Rifleman', 'Corporal', 2),
(4, 'Grenadier', 'Lance Corporal', 1),
(5, 'Rifleman', 'Lance Corporal', 1),
-- Bravo Team
(6, 'Team Leader', 'Sergeant', 1),
(7, 'Automatic Rifleman', 'Corporal', 2),
(8, 'Designated Marksman', 'Lance Corporal', 4),
(9, 'Rifleman', 'Lance Corporal', 1),
-- Machine Gun Team
(10, 'Machine Gun Squad Leader', 'Staff Sergeant', 1),
-- Charlie Team
(11, 'Machine Gunner', 'Corporal', 2),
(12, 'Assistant Gunner', 'Lance Corporal', 1),
-- Delta Team
(13, 'Machine Gunner', 'Corporal', 2),
(14, 'Assistant Gunner', 'Lance Corporal', 1);


-- Test data for teams
INSERT INTO teams (team_id, team_name, team_size) VALUES
(1, 'Alpha', 4),
(2, 'Bravo', 4),
(3, 'Charlie', 2),
(4, 'Delta', 2);

-- Test data for team_members
INSERT INTO team_members (team_id, member_id) VALUES
-- Alpha Team
(1, 2),
(1, 3),
(1, 4),
(1, 5),
-- Bravo Team
(2, 6),
(2, 7),
(2, 8),
(2, 9),
-- Charlie Team
(3, 11),
(3, 12),
-- Delta Team
(4, 13),
(4, 14);

-- Test data for groups
INSERT INTO groups (group_id, group_name, group_size, group_nationality) VALUES
(1, 'Ranger Rifle Squad', 9, 'United States of America'),
(2, 'Ranger Machine Gun Squad', 5, 'United States of America');


-- Test data for group_members
INSERT INTO group_members (group_id, member_id, team_id) VALUES
(1, 1, NULL),      -- Squad Leader
(1, NULL, 1),      -- Alpha Team
(1, NULL, 2),      -- Bravo Team
(2, 10, NULL),     -- Machine Gun Squad Leader
(2, NULL, 3),      -- Charlie Team
(2, NULL, 4);      -- Delta Team


-- +goose Down
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS members;
DROP TABLE IF EXISTS weapons;