-- +goose Up
-- Add new weapon type
INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber) VALUES
(6, 'M240B', 'Medium Machine Gun', '7.62mm');

-- Members
INSERT INTO members (member_id, member_role, member_rank) VALUES
(44, 'Squad Leader', 'Staff Sergeant'),
-- Alpha Team
(45, 'Team Leader', 'Sergeant'),
(46, 'Machine Gunner', 'Corporal'),
(47, 'Assistant Gunner', 'Lance Corporal'),
-- Bravo Team
(48, 'Team Leader', 'Sergeant'),
(49, 'Machine Gunner', 'Corporal'),
(50, 'Assistant Gunner', 'Lance Corporal');

-- Weapon assignments
INSERT INTO members_weapons (member_id, weapon_id) VALUES
(44, 5), -- Squad Leader - M27
(45, 5), -- Team Leader - M27
(46, 6), -- Machine Gunner - M240B
(47, 5), -- Assistant Gunner - M27
(48, 5), -- Team Leader - M27
(49, 6), -- Machine Gunner - M240B
(50, 5); -- Assistant Gunner - M27

-- Teams
INSERT INTO teams (team_id, team_name, team_size) VALUES
(8, 'Alpha Team', 3),
(9, 'Bravo Team', 3);

-- Team members
INSERT INTO team_members (team_id, member_id) VALUES
-- Alpha Team
(8, 45), (8, 46), (8, 47),
-- Bravo Team
(9, 48), (9, 49), (9, 50);

-- Group
INSERT INTO groups (group_id, group_name, group_size, group_nationality) VALUES
(4, 'Marine Weapons Squad', 7, 'United States of America');

-- Group members
INSERT INTO group_members (group_id, member_id, team_id) VALUES
(4, 44, NULL),     -- Squad Leader
(4, NULL, 8),      -- Alpha Team
(4, NULL, 9);      -- Bravo Team

-- +goose Down
DELETE FROM group_members WHERE group_id = 4;
DELETE FROM groups WHERE group_id = 4;
DELETE FROM team_members WHERE team_id IN (8, 9);
DELETE FROM teams WHERE team_id IN (8, 9);
DELETE FROM members_weapons WHERE member_id BETWEEN 44 AND 50;
DELETE FROM members WHERE member_id BETWEEN 44 AND 50;
DELETE FROM weapons WHERE weapon_id = 6;