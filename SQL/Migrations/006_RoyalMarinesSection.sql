-- +goose Up
-- Add new weapons
INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber) VALUES
(7, 'L85A3', 'Assault Rifle', '5.56mm'),
(8, 'L7A2', 'General Purpose Machine Gun', '7.62mm');

-- Members
INSERT INTO members (member_id, member_role, member_rank) VALUES
(51, 'Section Commander', 'Corporal'),
-- Alpha Fire Team
(52, 'Team Leader', 'Lance Corporal'),
(53, 'Gunner', 'Marine'),
(54, 'Rifleman', 'Marine'),
-- Bravo Fire Team
(55, 'Team Leader', 'Lance Corporal'),
(56, 'Gunner', 'Marine'),
(57, 'Rifleman', 'Marine');

-- Weapon assignments
INSERT INTO members_weapons (member_id, weapon_id) VALUES
(51, 7), -- Section Commander - L85A3
(52, 7), -- Team Leader - L85A3
(53, 8), -- Gunner - L7A2
(54, 7), -- Rifleman - L85A3
(55, 7), -- Team Leader - L85A3
(56, 8), -- Gunner - L7A2
(57, 7); -- Rifleman - L85A3

-- Teams
INSERT INTO teams (team_id, team_name, team_size) VALUES
(10, 'Alpha Fire Team', 3),
(11, 'Bravo Fire Team', 3);

-- Team members
INSERT INTO team_members (team_id, member_id) VALUES
-- Alpha Fire Team
(10, 52), (10, 53), (10, 54),
-- Bravo Fire Team
(11, 55), (11, 56), (11, 57);

-- Group
INSERT INTO groups (group_id, group_name, group_size, group_nationality) VALUES
(5, 'Royal Marines Section', 7, 'United Kingdom');

-- Group members
INSERT INTO group_members (group_id, member_id, team_id) VALUES
(5, 51, NULL),     -- Section Commander
(5, NULL, 10),     -- Alpha Fire Team
(5, NULL, 11);     -- Bravo Fire Team

-- +goose Down
DELETE FROM group_members WHERE group_id = 5;
DELETE FROM groups WHERE group_id = 5;
DELETE FROM team_members WHERE team_id IN (10, 11);
DELETE FROM teams WHERE team_id IN (10, 11);
DELETE FROM members_weapons WHERE member_id BETWEEN 51 AND 57;
DELETE FROM members WHERE member_id BETWEEN 51 AND 57;
DELETE FROM weapons WHERE weapon_id IN (7, 8);