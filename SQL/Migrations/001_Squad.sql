-- +goose Up
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS members_weapons;
DROP TABLE IF EXISTS members;
DROP TABLE IF EXISTS weapons;

CREATE TABLE weapons (
    weapon_id INTEGER PRIMARY KEY,
    weapon_name TEXT,
    weapon_type TEXT,
    weapon_caliber TEXT
);

CREATE TABLE members_weapons (
    member_id INTEGER,
    weapon_id INTEGER,
    PRIMARY KEY (member_id, weapon_id),
    FOREIGN KEY (member_id) REFERENCES members(member_id),
    FOREIGN KEY (weapon_id) REFERENCES weapons(weapon_id)
);

CREATE TABLE members (
    member_id INTEGER PRIMARY KEY,
    member_role TEXT,
    member_rank TEXT
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
(4, 'M110', 'Designated Marksman Rifle', '7.62mm'),
(5, 'M27 IAR', 'Infantry Automatic Rifle', '5.56mm');

-- Test data for members (US Army Ranger Rifle Squad)
INSERT INTO members (member_id, member_role, member_rank) VALUES
(1, 'Squad Leader', 'Staff Sergeant'),
-- Alpha Team
(2, 'Team Leader', 'Sergeant'),
(3, 'Automatic Rifleman', 'Corporal'),
(4, 'Grenadier', 'Lance Corporal'),
(5, 'Rifleman', 'Lance Corporal'),
-- Bravo Team
(6, 'Team Leader', 'Sergeant'),
(7, 'Automatic Rifleman', 'Corporal'),
(8, 'Designated Marksman', 'Lance Corporal'),
(9, 'Rifleman', 'Lance Corporal'),
-- Machine Gun Team
(10, 'Machine Gun Squad Leader', 'Staff Sergeant'),
-- Charlie Team
(11, 'Machine Gunner', 'Corporal'),
(12, 'Assistant Gunner', 'Lance Corporal'),
-- Delta Team
(13, 'Machine Gunner', 'Corporal'),
(14, 'Assistant Gunner', 'Lance Corporal'),
-- Marine Rifle Squad
(15, 'Squad Leader', 'Sergeant'),
-- 1st Fire Team
(16, 'Team Leader', 'Corporal'),
(17, 'Automatic Rifleman', 'Lance Corporal'),
(18, 'Grenadier', 'Lance Corporal'),
(19, 'Rifleman', 'Lance Corporal'),
-- 2nd Fire Team
(20, 'Team Leader', 'Corporal'),
(21, 'Automatic Rifleman', 'Lance Corporal'),
(22, 'Grenadier', 'Lance Corporal'),
(23, 'Rifleman', 'Lance Corporal'),
-- 3rd Fire Team
(24, 'Team Leader', 'Corporal'),
(25, 'Automatic Rifleman', 'Lance Corporal'),
(26, 'Grenadier', 'Lance Corporal'),
(27, 'Rifleman', 'Lance Corporal');

-- Weapon assignments
INSERT INTO members_weapons (member_id, weapon_id) VALUES
-- Ranger Squad
(1, 1),  -- Squad Leader - M4A1
(2, 1),  -- Team Leader - M4A1
(3, 2),  -- Auto Rifleman - M249
(4, 1),  -- Grenadier - M4A1
(4, 3),  -- Grenadier - M320
(5, 1),  -- Rifleman - M4A1
(6, 1),  -- Team Leader - M4A1
(7, 2),  -- Auto Rifleman - M249
(8, 4),  -- Designated Marksman - M110
(9, 1),  -- Rifleman - M4A1
(10, 1), -- MG Squad Leader - M4A1
(11, 2), -- Machine Gunner - M249
(12, 1), -- Assistant Gunner - M4A1
(13, 2), -- Machine Gunner - M249
(14, 1), -- Assistant Gunner - M4A1
-- Marine Squad
(15, 5), -- Squad Leader - M27
(16, 5), -- Team Leader - M27
(17, 5), -- Auto Rifleman - M27
(18, 5), -- Grenadier - M27
(18, 3), -- Grenadier - M320
(19, 5), -- Rifleman - M27
(20, 5), -- Team Leader - M27
(21, 5), -- Auto Rifleman - M27
(22, 5), -- Grenadier - M27
(22, 3), -- Grenadier - M320
(23, 5), -- Rifleman - M27
(24, 5), -- Team Leader - M27
(25, 5), -- Auto Rifleman - M27
(26, 5), -- Grenadier - M27
(26, 3), -- Grenadier - M320
(27, 5); -- Rifleman - M27

-- Test data for teams
INSERT INTO teams (team_id, team_name, team_size) VALUES
(1, 'Alpha', 4),
(2, 'Bravo', 4),
(3, 'Charlie', 2),
(4, 'Delta', 2),
(5, '1st Fire Team', 4),
(6, '2nd Fire Team', 4),
(7, '3rd Fire Team', 4);

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
(4, 14),
-- 1st Fire Team
(5, 16),
(5, 17),
(5, 18),
(5, 19),
-- 2nd Fire Team
(6, 20),
(6, 21),
(6, 22),
(6, 23),
-- 3rd Fire Team
(7, 24),
(7, 25),
(7, 26),
(7, 27);

-- Test data for groups
INSERT INTO groups (group_id, group_name, group_size, group_nationality) VALUES
(1, 'Ranger Rifle Squad', 9, 'United States of America'),
(2, 'Ranger Machine Gun Squad', 5, 'United States of America'),
(3, 'Marine Rifle Squad', 13, 'United States of America');

-- Test data for group_members
INSERT INTO group_members (group_id, member_id, team_id) VALUES
(1, 1, NULL),      -- Ranger Squad Leader
(1, NULL, 1),      -- Alpha Team
(1, NULL, 2),      -- Bravo Team
(2, 10, NULL),     -- Machine Gun Squad Leader
(2, NULL, 3),      -- Charlie Team
(2, NULL, 4),      -- Delta Team
(3, 15, NULL),     -- Marine Squad Leader
(3, NULL, 5),      -- 1st Fire Team
(3, NULL, 6),      -- 2nd Fire Team
(3, NULL, 7);      -- 3rd Fire Team

-- +goose Down
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS members_weapons;
DROP TABLE IF EXISTS members;
DROP TABLE IF EXISTS weapons;