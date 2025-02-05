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


// ... existing code until weapons INSERT ...

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
(9, 'Rifleman', 'Lance Corporal', 1);

-- Test data for teams
INSERT INTO teams (team_id, team_name, team_size) VALUES
(1, 'Alpha', 4),
(2, 'Bravo', 4);

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
(2, 9);

-- Test data for groups
INSERT INTO groups (group_id, group_name, group_size, group_nationality) VALUES
(1, 'Ranger Rifle Squad', 9, 'United States of America');

-- Test data for group_members
INSERT INTO group_members (group_id, member_id, team_id) VALUES
(1, 1, NULL),      -- Squad Leader
(1, NULL, 1),      -- Alpha Team
(1, NULL, 2);      -- Bravo Team

-- Additional test queries
-- 4. Show full squad composition with hierarchy
WITH squad_composition AS (
    -- Get individual members (Squad Leader)
    SELECT 
        g.group_name,
        m.member_role,
        m.member_rank,
        NULL as team_name,
        w.weapon_name
    FROM groups g
    JOIN group_members gm ON g.group_id = gm.group_id
    JOIN members m ON gm.member_id = m.member_id
    JOIN weapons w ON m.weapon_id = w.weapon_id
    WHERE gm.team_id IS NULL

    UNION ALL

    -- Get team members
    SELECT 
        g.group_name,
        m.member_role,
        m.member_rank,
        t.team_name,
        w.weapon_name
    FROM groups g
    JOIN group_members gm ON g.group_id = gm.group_id
    JOIN teams t ON gm.team_id = t.team_id
    JOIN team_members tm ON t.team_id = tm.team_id
    JOIN members m ON tm.member_id = m.member_id
    JOIN weapons w ON m.weapon_id = w.weapon_id
)
SELECT * FROM squad_composition ORDER BY 
    CASE 
        WHEN member_role = 'Squad Leader' THEN 1
        WHEN team_name = 'Alpha' THEN 2
        WHEN team_name = 'Bravo' THEN 3
        ELSE 4
    END,
    CASE 
        WHEN member_role = 'Team Leader' THEN 1
        WHEN member_role = 'Automatic Rifleman' THEN 2
        ELSE 3
    END;

-- 5. Weapon distribution analysis
SELECT 
    w.weapon_type,
    COUNT(*) as count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM members), 2) as percentage
FROM weapons w
JOIN members m ON w.weapon_id = m.weapon_id
GROUP BY w.weapon_type
ORDER BY count DESC;

-- 6. Team equipment summary
SELECT 
    t.team_name,
    COUNT(DISTINCT m.member_id) as member_count,
    COUNT(DISTINCT w.weapon_type) as unique_weapon_types,
    GROUP_CONCAT(DISTINCT w.weapon_type) as weapon_types
FROM teams t
JOIN team_members tm ON t.team_id = tm.team_id
JOIN members m ON tm.member_id = m.member_id
JOIN weapons w ON m.weapon_id = w.weapon_id
GROUP BY t.team_name;


-- +goose Down
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS members;
DROP TABLE IF EXISTS weapons;