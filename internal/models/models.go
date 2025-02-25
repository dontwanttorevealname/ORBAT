package models

import (
	"database/sql"
)

// Group represents a military group
type Group struct {
	ID          int
	Name        string
	Size        int
	Nationality string
}

// Weapon represents a weapon type
type Weapon struct {
	ID       int
	Name     string
	Type     string
	Caliber  string
	ImageURL sql.NullString
}

// Member represents a member of a group or team
type Member struct {
	ID      int
	Role    string
	Rank    string
	Weapons []Weapon
}

// Team represents a team within a group
type Team struct {
	ID      int
	Name    string
	Size    int
	Members []Member
}

// WeaponUser represents a user of a specific weapon
type WeaponUser struct {
	Role     string
	Rank     string
	TeamName string
}

// WeaponGroupUsers represents groups using a specific weapon
type WeaponGroupUsers struct {
	GroupName    string
	GroupID      int
	Nationality  string
	Users        []WeaponUser
}

// WeaponDetails represents detailed information about a weapon
type WeaponDetails struct {
	Weapon       Weapon
	TotalUsers   int
	Groups       []WeaponGroupUsers
	CountryCount int
	Countries    []string
}

// Vehicle represents a military vehicle
type Vehicle struct {
	ID        string
	Name      string
	Type      string
	Armament  string
	ImageURL  sql.NullString
	Crew      []Member
}

// GroupDetails represents detailed information about a group
type GroupDetails struct {
	ID            string
	Name          string
	Size          int
	Nationality   string
	DirectMembers []Member
	Teams         []Team
	Vehicles      []Vehicle
}

// VehicleDetails represents detailed information about a vehicle
type VehicleDetails struct {
	Vehicle      Vehicle
	Groups       []VehicleGroupUsers
	TotalUsers   int
	CountryCount int
	Countries    []string
}

// VehicleGroupUsers represents groups using a specific vehicle
type VehicleGroupUsers struct {
	GroupID     int
	GroupName   string
	Nationality string
	Members     []VehicleMember
}

// VehicleMember represents a crew member of a vehicle
type VehicleMember struct {
	Role string
	Rank string
}

// CountryDetails represents detailed information about a country
type CountryDetails struct {
	Name     string
	Groups   []Group
	Weapons  []WeaponUsage
	Vehicles []VehicleUsage
}

// WeaponUsage represents usage statistics for a weapon
type WeaponUsage struct {
	Weapon
	UserCount int
}

// VehicleUsage represents usage statistics for a vehicle
type VehicleUsage struct {
	Vehicle
	InstanceCount int
} 