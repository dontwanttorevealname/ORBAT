package main

import (
    "database/sql"
    "fmt"
    "html/template"
    "net/http"
    "os"
    _ "github.com/tursodatabase/libsql-client-go/libsql"
    "github.com/joho/godotenv"
)

type Group struct {
    ID           int    
    Name         string
    Size         int
    Nationality  string
}

type Weapon struct {
    ID      int
    Name    string
    Type    string
    Caliber string
}

type Member struct {
    ID     int
    Role   string
    Rank   string
    Weapon Weapon
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

func main() {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        fmt.Printf("Error loading .env file: %v\n", err)
        return
    }

    // Connect to the database using environment variable
    db, err := sql.Open("libsql", os.Getenv("DATABASE_URL"))
    if err != nil {
        fmt.Printf("Failed to connect to database: %v\n", err)
        return
    }
    defer db.Close()

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

        tmpl := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Military Order of Battle</title>
            <style>
                body {
                    font-family: Arial, sans-serif;
                    max-width: 800px;
                    margin: 0 auto;
                    padding: 20px;
                }
                h1 {
                    color: #333;
                }
                .group {
                    margin: 10px 0;
                    padding: 10px;
                    border: 1px solid #ddd;
                    border-radius: 4px;
                }
                a {
                    color: #0066cc;
                    text-decoration: none;
                }
                a:hover {
                    text-decoration: underline;
                }
            </style>
        </head>
        <body>
            <h1>Military Order of Battle</h1>
            {{range .}}
            <div class="group">
                <a href="/group/{{.ID}}">{{.Name}}</a> - {{.Nationality}} (Size: {{.Size}})
            </div>
            {{end}}
        </body>
        </html>`

        t := template.Must(template.New("groups").Parse(tmpl))
        t.Execute(w, groups)
    })

    // Handler for group details
    http.HandleFunc("/group/", func(w http.ResponseWriter, r *http.Request) {
        id := r.URL.Path[len("/group/"):]
        group, err := getGroupDetails(db, id)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        tmpl := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>{{.Name}} - Details</title>
            <style>
                body {
                    font-family: Arial, sans-serif;
                    max-width: 800px;
                    margin: 0 auto;
                    padding: 20px;
                }
                .section {
                    margin: 20px 0;
                    padding: 15px;
                    border: 1px solid #ddd;
                    border-radius: 4px;
                }
                .member, .team {
                    margin: 10px 0;
                    padding: 10px;
                    background: #f5f5f5;
                    border-radius: 4px;
                }
                h2, h3 {
                    color: #333;
                }
            </style>
        </head>
        <body>
            <h1>{{.Name}} ({{.Nationality}})</h1>
            <p>Total Size: {{.Size}}</p>
            
            {{if .DirectMembers}}
            <div class="section">
                <h2>Direct Members</h2>
                {{range .DirectMembers}}
                <div class="member">
                    <h3>{{.Role}} - {{.Rank}}</h3>
                    <p>Weapon: {{.Weapon.Name}} ({{.Weapon.Type}}, {{.Weapon.Caliber}})</p>
                </div>
                {{end}}
            </div>
            {{end}}

            {{if .Teams}}
            <div class="section">
                <h2>Teams</h2>
                {{range .Teams}}
                <div class="team">
                    <h3>{{.Name}} (Size: {{.Size}})</h3>
                    {{range .Members}}
                    <div class="member">
                        <h4>{{.Role}} - {{.Rank}}</h4>
                        <p>Weapon: {{.Weapon.Name}} ({{.Weapon.Type}}, {{.Weapon.Caliber}})</p>
                    </div>
                    {{end}}
                </div>
                {{end}}
            </div>
            {{end}}
            
            <p><a href="/">Back to Groups</a></p>
        </body>
        </html>`

        t := template.Must(template.New("group").Parse(tmpl))
        t.Execute(w, group)
    })

    fmt.Println("Server starting on http://localhost:8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Printf("Server error: %v\n", err)
    }
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

func getGroupDetails(db *sql.DB, groupID string) (GroupDetails, error) {
    var group GroupDetails
    
    // Get basic group info
    err := db.QueryRow(`
        SELECT group_id, group_name, group_size, group_nationality 
        FROM groups WHERE group_id = ?`, groupID).Scan(&group.ID, &group.Name, &group.Size, &group.Nationality)
    if err != nil {
        return group, err
    }

    // Get direct members
    memberRows, err := db.Query(`
        SELECT m.member_id, m.member_role, m.member_rank, 
               w.weapon_id, w.weapon_name, w.weapon_type, w.weapon_caliber
        FROM group_members gm
        JOIN members m ON gm.member_id = m.member_id
        JOIN weapons w ON m.weapon_id = w.weapon_id
        WHERE gm.group_id = ? AND gm.team_id IS NULL`, groupID)
    if err != nil {
        return group, err
    }
    defer memberRows.Close()

    for memberRows.Next() {
        var m Member
        var w Weapon
        err := memberRows.Scan(&m.ID, &m.Role, &m.Rank, &w.ID, &w.Name, &w.Type, &w.Caliber)
        if err != nil {
            return group, err
        }
        m.Weapon = w
        group.DirectMembers = append(group.DirectMembers, m)
    }

    // Get teams and their members
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

        // Get team members
        teamMemberRows, err := db.Query(`
            SELECT m.member_id, m.member_role, m.member_rank,
                   w.weapon_id, w.weapon_name, w.weapon_type, w.weapon_caliber
            FROM team_members tm
            JOIN members m ON tm.member_id = m.member_id
            JOIN weapons w ON m.weapon_id = w.weapon_id
            WHERE tm.team_id = ?`, team.ID)
        if err != nil {
            return group, err
        }
        defer teamMemberRows.Close()

        for teamMemberRows.Next() {
            var m Member
            var w Weapon
            err := teamMemberRows.Scan(&m.ID, &m.Role, &m.Rank, &w.ID, &w.Name, &w.Type, &w.Caliber)
            if err != nil {
                return group, err
            }
            m.Weapon = w
            team.Members = append(team.Members, m)
        }

        group.Teams = append(group.Teams, team)
    }

    return group, nil
}	