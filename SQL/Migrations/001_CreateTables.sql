-- +goose Up
CREATE TABLE weapons (
    weapon_id INTEGER PRIMARY KEY,
    weapon_name TEXT,
    weapon_type TEXT,
    weapon_caliber TEXT
);

CREATE TABLE members (
    member_id INTEGER PRIMARY KEY,
    member_role TEXT,
    member_rank TEXT
);

CREATE TABLE members_weapons (
    member_id INTEGER,
    weapon_id INTEGER,
    PRIMARY KEY (member_id, weapon_id),
    FOREIGN KEY (member_id) REFERENCES members(member_id),
    FOREIGN KEY (weapon_id) REFERENCES weapons(weapon_id)
);

CREATE TABLE teams (
    team_id INTEGER PRIMARY KEY,
    team_name TEXT,
    team_size INTEGER
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
    group_size INTEGER,
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

-- +goose Down
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS members_weapons;
DROP TABLE IF EXISTS members;
DROP TABLE IF EXISTS weapons;