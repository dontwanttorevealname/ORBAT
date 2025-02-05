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
    weapon_id TEXT,
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
(1, 'M4A1', 'Rifle', '5.56mm'),
(2, 'Glock 17', 'Pistol', '9mm'),
(3, 'M249', 'LMG', '5.56mm');

-- Test data for members
INSERT INTO members (member_id, member_role, member_rank, weapon_id) VALUES
(1, 'Squad Leader', 'Sergeant', '1'),
(2, 'Medic', 'Corporal', '2'),
(3, 'Machine Gunner', 'Private', '3');

-- Test data for teams
INSERT INTO teams (team_id, team_name, team_size) VALUES
(1, 'Alpha', 2),
(2, 'Bravo', 1);

-- Test data for team_members
INSERT INTO team_members (team_id, member_id) VALUES
(1, 1),
(1, 2),
(2, 3);

-- Test data for groups
INSERT INTO groups (group_id, group_name, group_size, group_nationality) VALUES
(1, 'First Platoon', 3, 'US'),
(2, 'Second Platoon', 2, 'UK');

-- Test data for group_members
INSERT INTO group_members (group_id, member_id, team_id) VALUES
(1, 1, NULL),    -- Individual member in group 1
(1, NULL, 2),    -- Team 2 in group 1
(2, NULL, 1);    -- Team 1 in group 2

-- Test queries
-- 1. Show all members and their weapons
SELECT m.member_id, m.member_role, m.member_rank, w.weapon_name
FROM members m
JOIN weapons w ON m.weapon_id = w.weapon_id;

-- 2. Show team composition
SELECT t.team_name, m.member_role, m.member_rank
FROM teams t
JOIN team_members tm ON t.team_id = tm.team_id
JOIN members m ON tm.member_id = m.member_id;

-- 3. Show group composition (both individual members and teams)
SELECT 
    g.group_name,
    m.member_role as individual_member,
    t.team_name as team_name
FROM groups g
LEFT JOIN group_members gm ON g.group_id = gm.group_id
LEFT JOIN members m ON gm.member_id = m.member_id
LEFT JOIN teams t ON gm.team_id = t.team_id;

-- +goose Down
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS members;
DROP TABLE IF EXISTS weapons;