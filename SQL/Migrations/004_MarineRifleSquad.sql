-- +goose Up
-- Add new weapon type
INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber) VALUES
(5, 'M27 IAR', 'Infantry Automatic Rifle', '5.56mm');

-- Members
INSERT INTO members (member_id, member_role, member_rank) VALUES
(31, 'Squad Leader', 'Sergeant'),
-- 1st Fire Team
(32, 'Team Leader', 'Corporal'),
(33, 'Automatic Rifleman', 'Lance Corporal'),
(34, 'Grenadier', 'Lance Corporal'),
(35, 'Rifleman', 'Lance Corporal'),
-- 2nd Fire Team
(36, 'Team Leader', 'Corporal'),
(37, 'Automatic Rifleman', 'Lance Corporal'),
(38, 'Grenadier', 'Lance Corporal'),
(39, 'Rifleman', 'Lance Corporal'),
-- 3rd Fire Team
(40, 'Team Leader', 'Corporal'),
(41, 'Automatic Rifleman', 'Lance Corporal'),
(42, 'Grenadier', 'Lance Corporal'),
(43, 'Rifleman', 'Lance Corporal');

-- Weapon assignments
INSERT INTO members_weapons (member_id, weapon_id) VALUES
(31, 5), -- Squad Leader - M27
(32, 5), -- Team Leader - M27
(33, 5), -- Auto Rifleman - M27
(34, 5), -- Grenadier - M27
(34, 3), -- Grenadier - M320
(35, 5), -- Rifleman - M27
(36, 5), -- Team Leader - M27
(37, 5), -- Auto Rifleman - M27
(38, 5), -- Grenadier - M27
(38, 3), -- Grenadier - M320
(39, 5), -- Rifleman - M27
(40, 5), -- Team Leader - M27
(41, 5), -- Auto Rifleman - M27
(42, 5), -- Grenadier - M27
(42, 3), -- Grenadier - M320
(43, 5); -- Rifleman - M27

-- Teams
INSERT INTO teams (team_id, team_name, team_size) VALUES
(5, '1st Fire Team', 4),
(6, '2nd Fire Team', 4),
(7, '3rd Fire Team', 4);

-- Team members
INSERT INTO team_members (team_id, member_id) VALUES
-- 1st Fire Team
(5, 32), (5, 33), (5, 34), (5, 35),
-- 2nd Fire Team
(6, 36), (6, 37), (6, 38), (6, 39),
-- 3rd Fire Team
(7, 40), (7, 41), (7, 42), (7, 43);

-- Group
INSERT INTO groups (group_id, group_name, group_size, group_nationality) VALUES
(3, 'Marine Rifle Squad', 13, 'United States of America');

-- Group members
INSERT INTO group_members (group_id, member_id, team_id) VALUES
(3, 31, NULL),     -- Squad Leader
(3, NULL, 5),      -- 1st Fire Team
(3, NULL, 6),      -- 2nd Fire Team
(3, NULL, 7);      -- 3rd Fire Team

-- +goose Down
DELETE FROM group_members WHERE group_id = 3;
DELETE FROM groups WHERE group_id = 3;
DELETE FROM team_members WHERE team_id IN (5, 6, 7);
DELETE FROM teams WHERE team_id IN (5, 6, 7);
DELETE FROM members_weapons WHERE member_id BETWEEN 31 AND 43;
DELETE FROM members WHERE member_id BETWEEN 31 AND 43;
DELETE FROM weapons WHERE weapon_id = 5;