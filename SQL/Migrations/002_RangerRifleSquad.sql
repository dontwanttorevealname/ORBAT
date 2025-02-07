-- +goose Up
-- Add weapons
INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber) VALUES
(1, 'M4A1', 'Assault Rifle', '5.56mm'),
(2, 'M249', 'Light Machine Gun', '5.56mm'),
(3, 'M320', 'Grenade Launcher', '40mm'),
(4, 'M110', 'Designated Marksman Rifle', '7.62mm');

-- Members
INSERT INTO members (member_id, member_role, member_rank) VALUES
(1, 'Squad Leader', 'Staff Sergeant'),
-- Alpha Team
(2, 'Team Leader', 'Sergeant'),
(3, 'Automatic Rifleman', 'Specialist'),
(4, 'Grenadier', 'Specialist'),
(5, 'Rifleman', 'Private First Class'),
-- Bravo Team
(6, 'Team Leader', 'Sergeant'),
(7, 'Automatic Rifleman', 'Specialist'),
(8, 'Grenadier', 'Specialist'),
(9, 'Rifleman', 'Private First Class');

-- Weapon assignments
INSERT INTO members_weapons (member_id, weapon_id) VALUES
(1, 1), -- Squad Leader - M4A1
(2, 1), -- Team Leader - M4A1
(3, 2), -- Auto Rifleman - M249
(4, 1), -- Grenadier - M4A1
(4, 3), -- Grenadier - M320
(5, 1), -- Rifleman - M4A1
(6, 1), -- Team Leader - M4A1
(7, 2), -- Auto Rifleman - M249
(8, 1), -- Grenadier - M4A1
(8, 3), -- Grenadier - M320
(9, 4); -- Rifleman - M110

-- Teams
INSERT INTO teams (team_id, team_name, team_size) VALUES
(1, 'Alpha Team', 4),
(2, 'Bravo Team', 4);

-- Team members
INSERT INTO team_members (team_id, member_id) VALUES
-- Alpha Team
(1, 2), (1, 3), (1, 4), (1, 5),
-- Bravo Team
(2, 6), (2, 7), (2, 8), (2, 9);

-- Group
INSERT INTO groups (group_id, group_name, group_size, group_nationality) VALUES
(1, 'Ranger Rifle Squad', 9, 'United States of America');

-- Group members
INSERT INTO group_members (group_id, member_id, team_id) VALUES
(1, 1, NULL),      -- Squad Leader
(1, NULL, 1),      -- Alpha Team
(1, NULL, 2);      -- Bravo Team

-- +goose Down
DELETE FROM group_members WHERE group_id = 1;
DELETE FROM groups WHERE group_id = 1;
DELETE FROM team_members WHERE team_id IN (1, 2);
DELETE FROM teams WHERE team_id IN (1, 2);
DELETE FROM members_weapons WHERE member_id BETWEEN 1 AND 9;
DELETE FROM members WHERE member_id BETWEEN 1 AND 9;
DELETE FROM weapons WHERE weapon_id BETWEEN 1 AND 4;