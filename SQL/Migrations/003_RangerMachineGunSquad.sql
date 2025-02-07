-- +goose Up
-- Members
INSERT INTO members (member_id, member_role, member_rank) VALUES
(10, 'Squad Leader', 'Staff Sergeant'),
-- Alpha Team
(11, 'Team Leader', 'Sergeant'),
(12, 'Machine Gunner', 'Specialist'),
(13, 'Assistant Gunner', 'Specialist'),
-- Bravo Team
(14, 'Team Leader', 'Sergeant'),
(15, 'Machine Gunner', 'Specialist'),
(16, 'Assistant Gunner', 'Specialist');

-- Weapon assignments
INSERT INTO members_weapons (member_id, weapon_id) VALUES
(10, 1), -- Squad Leader - M4A1
(11, 1), -- Team Leader - M4A1
(12, 2), -- Machine Gunner - M249
(13, 1), -- Assistant Gunner - M4A1
(14, 1), -- Team Leader - M4A1
(15, 2), -- Machine Gunner - M249
(16, 1); -- Assistant Gunner - M4A1

-- Teams
INSERT INTO teams (team_id, team_name, team_size) VALUES
(3, 'Alpha Team', 3),
(4, 'Bravo Team', 3);

-- Team members
INSERT INTO team_members (team_id, member_id) VALUES
-- Alpha Team
(3, 11), (3, 12), (3, 13),
-- Bravo Team
(4, 14), (4, 15), (4, 16);

-- Group
INSERT INTO groups (group_id, group_name, group_size, group_nationality) VALUES
(2, 'Ranger Machine Gun Squad', 7, 'United States of America');

-- Group members
INSERT INTO group_members (group_id, member_id, team_id) VALUES
(2, 10, NULL),     -- Squad Leader
(2, NULL, 3),      -- Alpha Team
(2, NULL, 4);      -- Bravo Team

-- +goose Down
DELETE FROM group_members WHERE group_id = 2;
DELETE FROM groups WHERE group_id = 2;
DELETE FROM team_members WHERE team_id IN (3, 4);
DELETE FROM teams WHERE team_id IN (3, 4);
DELETE FROM members_weapons WHERE member_id BETWEEN 10 AND 16;
DELETE FROM members WHERE member_id BETWEEN 10 AND 16;