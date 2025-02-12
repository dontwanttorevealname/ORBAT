package main

import (
    "cloud.google.com/go/storage"
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "html/template"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"
    _ "github.com/tursodatabase/libsql-client-go/libsql"
    "github.com/joho/godotenv"
)

type Group struct {
    ID          int
    Name        string
    Size        int
    Nationality string
}

type Weapon struct {
    ID       int
    Name     string
    Type     string
    Caliber  string
    ImageURL sql.NullString  // Change from string to sql.NullString
}

type Member struct {
    ID      int
    Role    string
    Rank    string
    Weapons []Weapon
}

type Team struct {
    ID      int
    Name    string
    Size    int
    Members []Member
}

type GroupDetails struct {
    ID            int
    Name          string
    Size          int
    Nationality   string
    DirectMembers []Member
    Teams         []Team
}

type WeaponUser struct {
    Role     string
    Rank     string
    TeamName string
}

type WeaponGroupUsers struct {
    GroupName    string
    GroupID      int
    Nationality  string
    Users        []WeaponUser
}

type WeaponDetails struct {
    Weapon       Weapon
    TotalUsers   int
    Groups       []WeaponGroupUsers
    CountryCount int
    Countries    []string
}

type Vehicle struct {
    ID        int
    Name      string
    Type      string
    Armament  string
    ImageURL  sql.NullString
}

type VehicleDetails struct {
    Vehicle   Vehicle
    Groups    []VehicleGroupUsers
    TotalUsers int
    CountryCount int
    Countries []string
}

type VehicleGroupUsers struct {
    GroupID     int
    GroupName   string
    Nationality string
    Members     []VehicleMember
}

type VehicleMember struct {
    Role    string
    Rank    string
}


var (
    templates     *template.Template
    db           *sql.DB
    storageClient *storage.Client
    bucketName   string
    adminPassword = "adminpassword"
)

func init() {
    // Initialize templates with better error handling
    templatesDir := "templates"
    var err error
    templates, err = template.ParseGlob(filepath.Join(templatesDir, "*.html"))
    if err != nil {
        fmt.Printf("Fatal: Failed to parse templates: %v\n", err)
        os.Exit(1)
    }

    // Load environment variables but don't exit if .env is missing
    if err := godotenv.Load(); err != nil {
        fmt.Printf("Info: .env file not found, using environment variables\n")
    }

    // Initialize database connection with retry logic
    maxRetries := 5
    for i := 0; i < maxRetries; i++ {
        db, err = sql.Open("libsql", os.Getenv("DATABASE_URL"))
        if err == nil {
            // Test the connection
            if err = db.Ping(); err == nil {
                fmt.Printf("Successfully connected to database\n")
                break
            }
        }
        fmt.Printf("Attempt %d: Failed to connect to database: %v\n", i+1, err)
        if i < maxRetries-1 {
            time.Sleep(time.Second * 2)
        }
    }
    if err != nil {
        fmt.Printf("Fatal: Could not establish database connection after %d attempts\n", maxRetries)
        os.Exit(1)
    }

    // Initialize Google Cloud Storage
    ctx := context.Background()
    storageClient, err = storage.NewClient(ctx)
    if err != nil {
        fmt.Printf("Failed to create storage client: %v\n", err)
        os.Exit(1)
    }

    bucketName = os.Getenv("GCS_BUCKET_NAME")
    if bucketName == "" {
        fmt.Printf("GCS_BUCKET_NAME environment variable not set\n")
        os.Exit(1)
    }
}

func uploadImageToGCS(file io.Reader, filename string) (string, error) {
    ctx := context.Background()
    ctx, cancel := context.WithTimeout(ctx, time.Minute*10)
    defer cancel()

    bucket := storageClient.Bucket(bucketName)
    obj := bucket.Object(filename)

    writer := obj.NewWriter(ctx)
    if _, err := io.Copy(writer, file); err != nil {
        return "", err
    }
    if err := writer.Close(); err != nil {
        return "", err
    }

    // Make the object public
    if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
        return "", err
    }

    return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, filename), nil
}

func main() {
    // Ensure database and storage client are closed
    defer func() {
        if db != nil {
            db.Close()
        }
        if storageClient != nil {
            storageClient.Close()
        }
    }()

    // Handler for the root path - shows all groups
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            http.NotFound(w, r)
            return
        }

        groups, err := getGroups(db)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        if err := templates.ExecuteTemplate(w, "groups.html", groups); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    })

    // Handler for weapons list and weapon addition
	http.HandleFunc("/weapons", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// Parse multipart form with 10MB max memory
			if err := r.ParseMultipartForm(10 << 20); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
	
			name := r.FormValue("name")
			weaponType := r.FormValue("type")
			caliber := r.FormValue("caliber")
			replace := r.FormValue("replace") == "true"
			
			// Check if weapon with this name exists
			exists, existingID, err := weaponExists(db, name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
	
			if exists && !replace {
				// Return a special status code to indicate name conflict
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte("Weapon with this name already exists"))
				return
			}
			
			var imageURL sql.NullString
			// Handle image upload if present
			file, header, err := r.FormFile("image")
			if err == nil {
				defer file.Close()
				
				filename := fmt.Sprintf("weapons/%s-%d%s", 
					strings.ToLower(strings.ReplaceAll(name, " ", "-")),
					time.Now().Unix(),
					filepath.Ext(header.Filename))
				
				uploadedURL, err := uploadImageToGCS(file, filename)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				imageURL = sql.NullString{String: uploadedURL, Valid: true}
			}
	
			tx, err := db.Begin()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer tx.Rollback()
	
			if exists && replace {
				var updateQuery string
				var params []interface{}
	
				if imageURL.Valid {
					updateQuery = `
						UPDATE weapons 
						SET weapon_type = ?,
							weapon_caliber = ?,
							image_url = ?
						WHERE weapon_id = ?`
					params = []interface{}{weaponType, caliber, imageURL.String, existingID}
				} else {
					updateQuery = `
						UPDATE weapons 
						SET weapon_type = ?,
							weapon_caliber = ?
						WHERE weapon_id = ?`
					params = []interface{}{weaponType, caliber, existingID}
				}
	
				_, err = tx.Exec(updateQuery, params...)
			} else {
				_, err = tx.Exec(`
					INSERT INTO weapons (weapon_name, weapon_type, weapon_caliber, image_url)
					VALUES (?, ?, ?, ?)`, 
					name, weaponType, caliber, imageURL)
			}
	
			if err != nil {
				http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
				return
			}
	
			if err := tx.Commit(); err != nil {
				http.Error(w, fmt.Sprintf("Failed to commit transaction: %v", err), http.StatusInternalServerError)
				return
			}
	
			http.Redirect(w, r, "/weapons", http.StatusSeeOther)
			return
		}
	
		// GET request handling remains the same
		weapons, err := getWeapons(db)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to fetch weapons: %v", err), http.StatusInternalServerError)
			return
		}
	
		if err := templates.ExecuteTemplate(w, "weapons.html", weapons); err != nil {
			http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
			return
		}
	})


    // Handler for vehicles list
    http.HandleFunc("/vehicles", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            if err := r.ParseMultipartForm(10 << 20); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }

            name := r.FormValue("name")
            vehicleType := r.FormValue("type")
            armament := r.FormValue("armament")
            if armament == "" {
                armament = "None"
            }

            // Check for duplicate names
            var exists bool
            err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM vehicles WHERE vehicle_name = ?)", name).Scan(&exists)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            if exists && r.FormValue("replace") != "true" {
                w.WriteHeader(http.StatusConflict)
                return
            }

            tx, err := db.Begin()
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            defer tx.Rollback()

            var vehicleID int64
            if exists {
                // Update existing vehicle
                _, err = tx.Exec(`
                    UPDATE vehicles 
                    SET vehicle_type = ?, vehicle_armament = ?
                    WHERE vehicle_name = ?`,
                    vehicleType, armament, name)
                if err != nil {
                    http.Error(w, err.Error(), http.StatusInternalServerError)
                    return
                }

                err = tx.QueryRow("SELECT vehicle_id FROM vehicles WHERE vehicle_name = ?", name).Scan(&vehicleID)
                if err != nil {
                    http.Error(w, err.Error(), http.StatusInternalServerError)
                    return
                }
            } else {
                // Insert new vehicle
                result, err := tx.Exec(`
                    INSERT INTO vehicles (vehicle_name, vehicle_type, vehicle_armament)
                    VALUES (?, ?, ?)`,
                    name, vehicleType, armament)
                if err != nil {
                    http.Error(w, err.Error(), http.StatusInternalServerError)
                    return
                }

                vehicleID, err = result.LastInsertId()
                if err != nil {
                    http.Error(w, err.Error(), http.StatusInternalServerError)
                    return
                }
            }

            // Handle image upload
            file, header, err := r.FormFile("image")
            if err == nil {
                defer file.Close()

                // Upload image to GCS
                filename := fmt.Sprintf("vehicles/%d_%s", vehicleID, header.Filename)
                imageURL, err := uploadImageToGCS(file, filename)
                if err != nil {
                    http.Error(w, err.Error(), http.StatusInternalServerError)
                    return
                }

                // Update vehicle with image URL
                _, err = tx.Exec("UPDATE vehicles SET image_url = ? WHERE vehicle_id = ?", imageURL, vehicleID)
                if err != nil {
                    http.Error(w, err.Error(), http.StatusInternalServerError)
                    return
                }
            }

            if err := tx.Commit(); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            http.Redirect(w, r, "/vehicles", http.StatusSeeOther)
            return
        }

        vehicles, err := getVehicles(db)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        if err := templates.ExecuteTemplate(w, "vehicles.html", vehicles); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    })

    // Handler for vehicle details
    http.HandleFunc("/vehicle/", func(w http.ResponseWriter, r *http.Request) {
        pathParts := strings.Split(r.URL.Path, "/")
        if len(pathParts) < 3 {
            http.NotFound(w, r)
            return
        }

        id := pathParts[2]
        if len(pathParts) == 4 && pathParts[3] == "delete" {
            if r.Method != "POST" {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
            }

            if err := deleteVehicle(db, id); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            http.Redirect(w, r, "/vehicles", http.StatusSeeOther)
            return
        }

        details, err := getVehicleDetails(db, id)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        if err := templates.ExecuteTemplate(w, "vehicle_details.html", details); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    })

    // Handler for managing member weapons
    http.HandleFunc("/member/", func(w http.ResponseWriter, r *http.Request) {
        pathParts := strings.Split(r.URL.Path, "/")
        if len(pathParts) != 4 || pathParts[3] != "weapons" {
            http.NotFound(w, r)
            return
        }

        memberID := pathParts[2]

        if r.Method == "POST" {
            if err := r.ParseForm(); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }

            if err := updateMemberWeapons(db, memberID, r.Form["weapons[]"]); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
            return
        }

        weapons, err := getMemberWeaponsData(db, memberID)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(weapons)
    })

    // Handler for weapon details and deletion
    http.HandleFunc("/weapon/", func(w http.ResponseWriter, r *http.Request) {
        pathParts := strings.Split(r.URL.Path, "/")
        if len(pathParts) < 3 {
            http.NotFound(w, r)
            return
        }

        id := pathParts[2]
        if len(pathParts) == 4 && pathParts[3] == "delete" {
            if r.Method != "POST" {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
            }

            if err := deleteWeapon(db, id); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            http.Redirect(w, r, "/weapons", http.StatusSeeOther)
            return
        }

        details, err := getWeaponDetails(db, id)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        if err := templates.ExecuteTemplate(w, "weapon_details.html", details); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    })

    // Handler for group details and deletion
    http.HandleFunc("/group/", func(w http.ResponseWriter, r *http.Request) {
        pathParts := strings.Split(r.URL.Path, "/")
        if len(pathParts) < 3 {
            http.NotFound(w, r)
            return
        }

        id := pathParts[2]
        if len(pathParts) == 4 && pathParts[3] == "delete" {
            if r.Method != "POST" {
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                return
            }

            if err := deleteGroup(db, id); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }

        group, err := getGroupDetails(db, id)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        if err := templates.ExecuteTemplate(w, "group_details.html", group); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    })

    // Handler for adding new groups
    http.HandleFunc("/add_group", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            if err := r.ParseForm(); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }

            if err := handleAddGroup(db, r); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }

        weapons, err := getWeapons(db)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        data := struct {
            WeaponOptions template.JS
        }{
            WeaponOptions: template.JS(weaponsToJSON(weapons)),
        }

        if err := templates.ExecuteTemplate(w, "add_group.html", data); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    })

    // Add basic health check endpoint
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        // Check database connection
        if err := db.Ping(); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            fmt.Fprintf(w, "Database connection error: %v", err)
            return
        }
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "OK")
    })

    // Get port from environment variable
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    // Create a server with timeouts
    srv := &http.Server{
        Addr:         ":" + port,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start server with improved logging
    fmt.Printf("Server starting on port %s\n", port)
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        fmt.Printf("Fatal: Server error: %v\n", err)
        os.Exit(1)
    }
}

func weaponExists(db *sql.DB, name string) (bool, int, error) {
    var id int
    err := db.QueryRow("SELECT weapon_id FROM weapons WHERE weapon_name = ?", name).Scan(&id)
    if err == sql.ErrNoRows {
        return false, 0, nil
    }
    if err != nil {
        return false, 0, err
    }
    return true, id, nil
}


func getGroups(db *sql.DB) ([]Group, error) {
    rows, err := db.Query("SELECT group_id, group_name, group_size, group_nationality FROM groups")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var groups []Group
    for rows.Next() {
        var g Group
        if err := rows.Scan(&g.ID, &g.Name, &g.Size, &g.Nationality); err != nil {
            return nil, err
        }
        groups = append(groups, g)
    }
    return groups, nil
}

func getWeapons(db *sql.DB) ([]Weapon, error) {
    rows, err := db.Query("SELECT weapon_id, weapon_name, weapon_type, weapon_caliber, image_url FROM weapons ORDER BY weapon_name")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var weapons []Weapon
    for rows.Next() {
        var w Weapon
        if err := rows.Scan(&w.ID, &w.Name, &w.Type, &w.Caliber, &w.ImageURL); err != nil {
            return nil, err
        }
        weapons = append(weapons, w)
    }
    return weapons, nil
}


func getGroupDetails(db *sql.DB, groupID string) (GroupDetails, error) {
    var group GroupDetails
    
    err := db.QueryRow(`
        SELECT group_id, group_name, group_size, group_nationality 
        FROM groups WHERE group_id = ?`, groupID).Scan(&group.ID, &group.Name, &group.Size, &group.Nationality)
    if err != nil {
        return group, err
    }

    memberRows, err := db.Query(`
        SELECT DISTINCT m.member_id, m.member_role, m.member_rank
        FROM group_members gm
        JOIN members m ON gm.member_id = m.member_id
        WHERE gm.group_id = ? AND gm.team_id IS NULL`, groupID)
    if err != nil {
        return group, err
    }
    defer memberRows.Close()

    for memberRows.Next() {
        var m Member
        err := memberRows.Scan(&m.ID, &m.Role, &m.Rank)
        if err != nil {
            return group, err
        }

        weaponRows, err := db.Query(`
            SELECT w.weapon_id, w.weapon_name, w.weapon_type, w.weapon_caliber
            FROM members_weapons mw
            JOIN weapons w ON mw.weapon_id = w.weapon_id
            WHERE mw.member_id = ?`, m.ID)
        if err != nil {
            return group, err
        }
        defer weaponRows.Close()

        for weaponRows.Next() {
            var w Weapon
            err := weaponRows.Scan(&w.ID, &w.Name, &w.Type, &w.Caliber)
            if err != nil {
                return group, err
            }
            m.Weapons = append(m.Weapons, w)
        }

        group.DirectMembers = append(group.DirectMembers, m)
    }

    teamRows, err := db.Query(`
        SELECT DISTINCT t.team_id, t.team_name, t.team_size
        FROM group_members gm
        JOIN teams t ON gm.team_id = t.team_id
        WHERE gm.group_id = ?`, groupID)
    if err != nil {
        return group, err
    }
    defer teamRows.Close()

    for teamRows.Next() {
        var team Team
        err := teamRows.Scan(&team.ID, &team.Name, &team.Size)
        if err != nil {
            return group, err
        }

        teamMemberRows, err := db.Query(`
            SELECT DISTINCT m.member_id, m.member_role, m.member_rank
            FROM team_members tm
            JOIN members m ON tm.member_id = m.member_id
            WHERE tm.team_id = ?`, team.ID)
        if err != nil {
            return group, err
        }
        defer teamMemberRows.Close()

        for teamMemberRows.Next() {
            var m Member
            err := teamMemberRows.Scan(&m.ID, &m.Role, &m.Rank)
            if err != nil {
                return group, err
            }

            weaponRows, err := db.Query(`
                SELECT w.weapon_id, w.weapon_name, w.weapon_type, w.weapon_caliber
                FROM members_weapons mw
                JOIN weapons w ON mw.weapon_id = w.weapon_id
                WHERE mw.member_id = ?`, m.ID)
            if err != nil {
                return group, err
            }
            defer weaponRows.Close()

            for weaponRows.Next() {
                var w Weapon
                err := weaponRows.Scan(&w.ID, &w.Name, &w.Type, &w.Caliber)
                if err != nil {
                    return group, err
                }
                m.Weapons = append(m.Weapons, w)
            }

            team.Members = append(team.Members, m)
        }

        group.Teams = append(group.Teams, team)
    }

    return group, nil
}

func getWeaponDetails(db *sql.DB, weaponID string) (WeaponDetails, error) {
    var details WeaponDetails

    err := db.QueryRow(`
        SELECT weapon_id, weapon_name, weapon_type, weapon_caliber, image_url 
        FROM weapons WHERE weapon_id = ?`, weaponID).Scan(
        &details.Weapon.ID, &details.Weapon.Name, &details.Weapon.Type, &details.Weapon.Caliber, &details.Weapon.ImageURL)
    if err != nil {
        return details, err
    }

    rows, err := db.Query(`
        SELECT 
            g.group_id,
            g.group_name,
            g.group_nationality,
            m.member_role,
            m.member_rank,
            t.team_name
        FROM members_weapons mw
        JOIN members m ON mw.member_id = m.member_id
        LEFT JOIN (
            SELECT member_id, group_id, NULL as team_id 
            FROM group_members 
            WHERE team_id IS NULL
            UNION ALL
            SELECT tm.member_id, gm.group_id, tm.team_id
            FROM team_members tm
            JOIN group_members gm ON tm.team_id = gm.team_id
        ) membership ON m.member_id = membership.member_id
        LEFT JOIN groups g ON membership.group_id = g.group_id
        LEFT JOIN teams t ON membership.team_id = t.team_id
        WHERE mw.weapon_id = ?
        ORDER BY g.group_name, t.team_name`, weaponID)
    if err != nil {
        return details, err
    }
    defer rows.Close()

    groupMap := make(map[int]*WeaponGroupUsers)
    countryMap := make(map[string]bool)
    userCount := 0

    for rows.Next() {
        var groupID sql.NullInt64
        var groupName sql.NullString
        var nationality sql.NullString
        var role string
        var rank string
        var teamName sql.NullString

        err := rows.Scan(&groupID, &groupName, &nationality, &role, &rank, &teamName)
        if err != nil {
            return details, err
        }

        user := WeaponUser{
            Role: role,
            Rank: rank,
        }
        if teamName.Valid {
            user.TeamName = teamName.String
        }

        gID := -1
        gName := "Unassigned"
        nat := "Unknown"
        
        if groupID.Valid {
            gID = int(groupID.Int64)
            gName = groupName.String
            nat = nationality.String
            countryMap[nat] = true
        }

        if group, exists := groupMap[gID]; exists {
            group.Users = append(group.Users, user)
        } else {
            groupMap[gID] = &WeaponGroupUsers{
                GroupID:     gID,
                GroupName:   gName,
                Nationality: nat,
                Users:      []WeaponUser{user},
            }
        }
        userCount++
    }

    details.TotalUsers = userCount
    
    for country := range countryMap {
        details.Countries = append(details.Countries, country)
    }
    details.CountryCount = len(details.Countries)
    
    for gid, group := range groupMap {
        if gid != -1 {
            details.Groups = append(details.Groups, *group)
        }
    }
    if unassigned, exists := groupMap[-1]; exists {
        details.Groups = append(details.Groups, *unassigned)
    }

    return details, nil
}

func weaponsToJSON(weapons []Weapon) string {
    type jsonWeapon struct {
        ID      int    `json:"ID"`
        Name    string `json:"Name"`
        Type    string `json:"Type"`
        Caliber string `json:"Caliber"`
    }

    jsonWeapons := make([]jsonWeapon, len(weapons))
    for i, w := range weapons {
		jsonWeapons[i] = jsonWeapon{
            ID:      w.ID,
            Name:    w.Name,
            Type:    w.Type,
            Caliber: w.Caliber,
        }
    }

    jsonData, err := json.Marshal(jsonWeapons)
    if err != nil {
        return "[]"
    }
    return string(jsonData)
}

func handleAddGroup(db *sql.DB, r *http.Request) error {
    if err := r.ParseForm(); err != nil {
        return err
    }

    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Insert group
    result, err := tx.Exec(`
        INSERT INTO groups (group_name, group_nationality, group_size)
        VALUES (?, ?, 0)`,
        r.FormValue("name"),
        r.FormValue("nationality"))
    if err != nil {
        return err
    }

    groupID, err := result.LastInsertId()
    if err != nil {
        return err
    }

    // Handle direct members
    directMemberRoles := r.Form["directMembers_role[]"]
    directMemberRanks := r.Form["directMembers_rank[]"]
    
    for i := range directMemberRoles {
        // Insert member
        result, err := tx.Exec(`
            INSERT INTO members (member_role, member_rank)
            VALUES (?, ?)`,
            directMemberRoles[i], directMemberRanks[i])
        if err != nil {
            return err
        }

        memberID, err := result.LastInsertId()
        if err != nil {
            return err
        }

        // Link member to group
        _, err = tx.Exec(`
            INSERT INTO group_members (group_id, member_id)
            VALUES (?, ?)`,
            groupID, memberID)
        if err != nil {
            return err
        }

        // Add weapons for this member
        weaponKey := fmt.Sprintf("weapons_%d[]", i)
        for _, weaponID := range r.Form[weaponKey] {
            _, err = tx.Exec(`
                INSERT INTO members_weapons (member_id, weapon_id)
                VALUES (?, ?)`,
                memberID, weaponID)
            if err != nil {
                return err
            }
        }
    }

    // Handle teams
    teamNames := r.Form["team_name[]"]
    for teamIndex := range teamNames {
        // Create team
        result, err := tx.Exec(`
            INSERT INTO teams (team_name, team_size)
            VALUES (?, 0)`,
            teamNames[teamIndex])
        if err != nil {
            return err
        }

        teamID, err := result.LastInsertId()
        if err != nil {
            return err
        }

        // Link team to group
        _, err = tx.Exec(`
            INSERT INTO group_members (group_id, team_id)
            VALUES (?, ?)`,
            groupID, teamID)
        if err != nil {
            return err
        }

        // Get team member data
        teamMemberRoles := r.Form[fmt.Sprintf("team_%d_role[]", teamIndex)]
        teamMemberRanks := r.Form[fmt.Sprintf("team_%d_rank[]", teamIndex)]

        for memberIndex := range teamMemberRoles {
            // Insert team member
            result, err := tx.Exec(`
                INSERT INTO members (member_role, member_rank)
                VALUES (?, ?)`,
                teamMemberRoles[memberIndex],
                teamMemberRanks[memberIndex])
            if err != nil {
                return err
            }

            memberID, err := result.LastInsertId()
            if err != nil {
                return err
            }

            // Link member to team
            _, err = tx.Exec(`
                INSERT INTO team_members (team_id, member_id)
                VALUES (?, ?)`,
                teamID, memberID)
            if err != nil {
                return err
            }

            // Add weapons for this team member
            weaponKey := fmt.Sprintf("team_%d_weapons_%d[]", teamIndex, memberIndex)
            for _, weaponID := range r.Form[weaponKey] {
                _, err = tx.Exec(`
                    INSERT INTO members_weapons (member_id, weapon_id)
                    VALUES (?, ?)`,
                    memberID, weaponID)
                if err != nil {
                    return err
                }
            }
        }

        // Update team size
        _, err = tx.Exec(`
            UPDATE teams 
            SET team_size = (
                SELECT COUNT(*) 
                FROM team_members 
                WHERE team_id = ?
            )
            WHERE team_id = ?`,
            teamID, teamID)
        if err != nil {
            return err
        }
    }

    // Update group size
    if err := updateGroupSize(tx, int(groupID)); err != nil {
        return err
    }

    return tx.Commit()
}

func addMemberToGroup(tx *sql.Tx, groupID int, role, rank string, teamID *int64, r *http.Request) error {
    // Insert member
    result, err := tx.Exec(`
        INSERT INTO members (member_role, member_rank)
        VALUES (?, ?)`,
        role, rank)
    if err != nil {
        return err
    }

    memberID, err := result.LastInsertId()
    if err != nil {
        return err
    }

    // Link member to group/team
    if teamID != nil {
        _, err = tx.Exec(`
            INSERT INTO team_members (team_id, member_id)
            VALUES (?, ?)`,
            *teamID, memberID)
    } else {
        _, err = tx.Exec(`
            INSERT INTO group_members (group_id, member_id)
            VALUES (?, ?)`,
            groupID, memberID)
    }
    if err != nil {
        return err
    }

    // Add weapons for this member
    weaponIDs := r.Form[fmt.Sprintf("weapons[]")]
    for _, weaponID := range weaponIDs {
        _, err = tx.Exec(`
            INSERT INTO members_weapons (member_id, weapon_id)
            VALUES (?, ?)`,
            memberID, weaponID)
        if err != nil {
            return err
        }
    }

    return nil
}

func updateGroupSize(tx *sql.Tx, groupID int) error {
    _, err := tx.Exec(`
        UPDATE groups 
        SET group_size = (
            SELECT COUNT(DISTINCT m.member_id)
            FROM members m
            LEFT JOIN group_members gm ON m.member_id = gm.member_id
            LEFT JOIN team_members tm ON m.member_id = tm.member_id
            LEFT JOIN group_members gt ON tm.team_id = gt.team_id
            WHERE gm.group_id = ? OR gt.group_id = ?
        )
        WHERE group_id = ?`,
        groupID, groupID, groupID)
    return err
}

func deleteWeapon(db *sql.DB, weaponID string) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Get the image URL before deleting the weapon
    var imageURL sql.NullString
    err = tx.QueryRow("SELECT image_url FROM weapons WHERE weapon_id = ?", weaponID).Scan(&imageURL)
    if err != nil {
        return err
    }

    // Delete weapon associations first
    _, err = tx.Exec("DELETE FROM members_weapons WHERE weapon_id = ?", weaponID)
    if err != nil {
        return err
    }

    // Delete the weapon itself
    _, err = tx.Exec("DELETE FROM weapons WHERE weapon_id = ?", weaponID)
    if err != nil {
        return err
    }

    // If there was an image, delete it from GCS
    if imageURL.Valid && imageURL.String != "" {
        // Extract filename from URL
        // URL format: https://storage.googleapis.com/BUCKET_NAME/weapons/filename
        urlParts := strings.Split(imageURL.String, "/")
        // The object name includes the "weapons/" prefix
        objectName := strings.Join(urlParts[len(urlParts)-2:], "/")
        
        fmt.Printf("Attempting to delete object: %s from bucket: %s\n", objectName, bucketName)

        ctx := context.Background()
        ctx, cancel := context.WithTimeout(ctx, time.Second*10)
        defer cancel()

        // Delete the object from GCS
        err = storageClient.Bucket(bucketName).Object(objectName).Delete(ctx)
        if err != nil {
            fmt.Printf("Warning: Failed to delete image from storage: %v\n", err)
            fmt.Printf("URL: %s\n", imageURL.String)
            fmt.Printf("Object name: %s\n", objectName)
        } else {
            fmt.Printf("Successfully deleted object: %s\n", objectName)
        }
    }

    return tx.Commit()
}

func deleteGroup(db *sql.DB, groupID string) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. First get all member IDs (both direct and team members) associated with this group
    memberIDs := make(map[string]bool)

    // Get direct member IDs
    directRows, err := tx.Query(`
        SELECT member_id 
        FROM group_members 
        WHERE group_id = ? AND team_id IS NULL`, groupID)
    if err != nil {
        return err
    }
    defer directRows.Close()

    for directRows.Next() {
        var memberID string
        if err := directRows.Scan(&memberID); err != nil {
            return err
        }
        memberIDs[memberID] = true
    }

    // Get team member IDs
    teamMemberRows, err := tx.Query(`
        SELECT tm.member_id
        FROM team_members tm
        JOIN group_members gm ON tm.team_id = gm.team_id
        WHERE gm.group_id = ?`, groupID)
    if err != nil {
        return err
    }
    defer teamMemberRows.Close()

    for teamMemberRows.Next() {
        var memberID string
        if err := teamMemberRows.Scan(&memberID); err != nil {
            return err
        }
        memberIDs[memberID] = true
    }

    // 2. Delete all weapon associations for these members
    for memberID := range memberIDs {
        _, err = tx.Exec("DELETE FROM members_weapons WHERE member_id = ?", memberID)
        if err != nil {
            return err
        }
    }

    // 3. Delete team_members associations
    _, err = tx.Exec(`
        DELETE FROM team_members 
        WHERE team_id IN (
            SELECT team_id 
            FROM group_members 
            WHERE group_id = ? AND team_id IS NOT NULL
        )`, groupID)
    if err != nil {
        return err
    }

    // 4. Delete group_members associations
    _, err = tx.Exec("DELETE FROM group_members WHERE group_id = ?", groupID)
    if err != nil {
        return err
    }

    // 5. Delete all members
    for memberID := range memberIDs {
        _, err = tx.Exec("DELETE FROM members WHERE member_id = ?", memberID)
        if err != nil {
            return err
        }
    }

    // 6. Delete teams
    _, err = tx.Exec(`
        DELETE FROM teams 
        WHERE team_id IN (
            SELECT DISTINCT team_id 
            FROM group_members 
            WHERE group_id = ? AND team_id IS NOT NULL
        )`, groupID)
    if err != nil {
        return err
    }

    // 7. Finally delete the group itself
    _, err = tx.Exec("DELETE FROM groups WHERE group_id = ?", groupID)
    if err != nil {
        return err
    }

    return tx.Commit()
}

func getMemberWeaponsData(db *sql.DB, memberID string) (map[string]interface{}, error) {
    // Get all available weapons
    allWeapons, err := getWeapons(db)
    if err != nil {
        return nil, err
    }

    // Get member's current weapons
    rows, err := db.Query(`
        SELECT w.weapon_id, w.weapon_name, w.weapon_type, w.weapon_caliber
        FROM members_weapons mw
        JOIN weapons w ON mw.weapon_id = w.weapon_id
        WHERE mw.member_id = ?`, memberID)
    if err != nil {
        // If there's an error querying current weapons, still return all weapons
        // but with an empty current weapons array
        return map[string]interface{}{
            "all":     allWeapons,
            "current": []Weapon{},
        }, nil
    }
    defer rows.Close()

    var currentWeapons []Weapon
    for rows.Next() {
        var w Weapon
        if err := rows.Scan(&w.ID, &w.Name, &w.Type, &w.Caliber); err != nil {
            // If there's an error scanning a weapon, skip it and continue
            continue
        }
        currentWeapons = append(currentWeapons, w)
    }

    // Even if there are no current weapons, we still return a valid response
    return map[string]interface{}{
        "all":     allWeapons,
        "current": currentWeapons,
    }, nil
}

func updateMemberWeapons(db *sql.DB, memberID string, weaponIDs []string) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Remove all existing weapons for this member
    _, err = tx.Exec("DELETE FROM members_weapons WHERE member_id = ?", memberID)
    if err != nil {
        return err
    }

    // If no weapons were selected, just return after deleting existing weapons
    if len(weaponIDs) == 0 {
        return tx.Commit()
    }

    // Add new weapons
    for _, weaponID := range weaponIDs {
        // Verify the weapon exists before inserting
        var exists bool
        err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM weapons WHERE weapon_id = ?)", weaponID).Scan(&exists)
        if err != nil {
            return err
        }
        if !exists {
            continue // Skip weapons that don't exist
        }

        _, err = tx.Exec(
            "INSERT INTO members_weapons (member_id, weapon_id) VALUES (?, ?)",
            memberID, weaponID)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}

func getVehicles(db *sql.DB) ([]Vehicle, error) {
    rows, err := db.Query("SELECT vehicle_id, vehicle_name, vehicle_type, vehicle_armament, image_url FROM vehicles ORDER BY vehicle_name")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var vehicles []Vehicle
    for rows.Next() {
        var v Vehicle
        if err := rows.Scan(&v.ID, &v.Name, &v.Type, &v.Armament, &v.ImageURL); err != nil {
            return nil, err
        }
        vehicles = append(vehicles, v)
    }
    return vehicles, nil
}

func getVehicleDetails(db *sql.DB, vehicleID string) (VehicleDetails, error) {
    var details VehicleDetails

    err := db.QueryRow(`
        SELECT vehicle_id, vehicle_name, vehicle_type, vehicle_armament, image_url 
        FROM vehicles WHERE vehicle_id = ?`, vehicleID).Scan(
        &details.Vehicle.ID, &details.Vehicle.Name, &details.Vehicle.Type, 
        &details.Vehicle.Armament, &details.Vehicle.ImageURL)
    if err != nil {
        return details, err
    }

    rows, err := db.Query(`
        SELECT 
            g.group_id,
            g.group_name,
            g.group_nationality,
            m.member_role,
            m.member_rank
        FROM vehicle_members vm
        JOIN members m ON vm.member_id = m.member_id
        JOIN group_vehicles gv ON vm.vehicle_id = gv.vehicle_id
        JOIN groups g ON gv.group_id = g.group_id
        WHERE vm.vehicle_id = ?
        ORDER BY g.group_name`, vehicleID)
    if err != nil {
        return details, err
    }
    defer rows.Close()

    groupMap := make(map[int]*VehicleGroupUsers)
    countryMap := make(map[string]bool)
    userCount := 0

    for rows.Next() {
        var groupID int
        var groupName, nationality, role, rank string

        err := rows.Scan(&groupID, &groupName, &nationality, &role, &rank)
        if err != nil {
            return details, err
        }

        countryMap[nationality] = true
        userCount++

        member := VehicleMember{
            Role: role,
            Rank: rank,
        }

        if group, exists := groupMap[groupID]; exists {
            group.Members = append(group.Members, member)
        } else {
            groupMap[groupID] = &VehicleGroupUsers{
                GroupID:     groupID,
                GroupName:   groupName,
                Nationality: nationality,
                Members:     []VehicleMember{member},
            }
        }
    }

    details.TotalUsers = userCount
    details.CountryCount = len(countryMap)
    
    for country := range countryMap {
        details.Countries = append(details.Countries, country)
    }

    for _, group := range groupMap {
        details.Groups = append(details.Groups, *group)
    }

    return details, nil
}

func deleteVehicle(db *sql.DB, vehicleID string) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Get the image URL before deleting the vehicle
    var imageURL sql.NullString
    err = tx.QueryRow("SELECT image_url FROM vehicles WHERE vehicle_id = ?", vehicleID).Scan(&imageURL)
    if err != nil {
        return err
    }

    // Delete vehicle associations first
    _, err = tx.Exec("DELETE FROM vehicle_members WHERE vehicle_id = ?", vehicleID)
    if err != nil {
        return err
    }

    _, err = tx.Exec("DELETE FROM group_vehicles WHERE vehicle_id = ?", vehicleID)
    if err != nil {
        return err
    }

    // Delete the vehicle
    _, err = tx.Exec("DELETE FROM vehicles WHERE vehicle_id = ?", vehicleID)
    if err != nil {
        return err
    }

    // Delete image from GCS if it exists
    if imageURL.Valid && imageURL.String != "" {
        if err := deleteImageFromGCS(imageURL.String); err != nil {
            return err
        }
    }

    return tx.Commit()
}