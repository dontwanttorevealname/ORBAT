-- +goose Up
-- Add NOT NULL constraints
ALTER TABLE weapons 
ADD CONSTRAINT weapons_required_fields 
CHECK (
    weapon_name IS NOT NULL 
    AND weapon_type IS NOT NULL 
    AND weapon_caliber IS NOT NULL
);

ALTER TABLE groups 
ADD CONSTRAINT groups_required_fields 
CHECK (
    group_name IS NOT NULL 
    AND group_nationality IS NOT NULL
);

-- Add foreign key constraints with CASCADE
DROP TABLE IF EXISTS group_members;
CREATE TABLE group_members (
    group_id INTEGER NOT NULL,
    member_id INTEGER NOT NULL,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES members(member_id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, member_id)
);

-- +goose Down
ALTER TABLE weapons DROP CONSTRAINT weapons_required_fields;
ALTER TABLE groups DROP CONSTRAINT groups_required_fields;

DROP TABLE IF EXISTS group_members;
CREATE TABLE group_members (
    group_id INTEGER,
    member_id INTEGER,
    FOREIGN KEY (group_id) REFERENCES groups(group_id),
    FOREIGN KEY (member_id) REFERENCES members(member_id),
    PRIMARY KEY (group_id, member_id)
);