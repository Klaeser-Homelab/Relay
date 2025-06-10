package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	conn *sql.DB
}

func NewDatabase() (*Database, error) {
	// Get home directory for storing relay data
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	
	// Create .relay directory if it doesn't exist
	relayDir := filepath.Join(homeDir, ".relay")
	if err := os.MkdirAll(relayDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create relay directory: %w", err)
	}
	
	// Database file path
	dbPath := filepath.Join(relayDir, "relay.db")
	
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	db := &Database{conn: conn}
	
	// Initialize schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}
	
	return db, nil
}

func (db *Database) initSchema() error {
	// Create projects table
	projectsSchema := `
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		path TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_opened DATETIME,
		description TEXT
	);`
	
	if _, err := db.conn.Exec(projectsSchema); err != nil {
		return fmt.Errorf("failed to create projects table: %w", err)
	}
	
	// Create project_settings table
	settingsSchema := `
	CREATE TABLE IF NOT EXISTS project_settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER,
		setting_key TEXT NOT NULL,
		setting_value TEXT,
		FOREIGN KEY (project_id) REFERENCES projects(id),
		UNIQUE(project_id, setting_key)
	);`
	
	if _, err := db.conn.Exec(settingsSchema); err != nil {
		return fmt.Errorf("failed to create project_settings table: %w", err)
	}
	
	// Create active_project table (stores which project is currently active)
	activeProjectSchema := `
	CREATE TABLE IF NOT EXISTS active_project (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		project_id INTEGER,
		FOREIGN KEY (project_id) REFERENCES projects(id)
	);`
	
	if _, err := db.conn.Exec(activeProjectSchema); err != nil {
		return fmt.Errorf("failed to create active_project table: %w", err)
	}
	
	return nil
}

func (db *Database) AddProject(name, path string) error {
	query := `
	INSERT INTO projects (name, path, created_at, last_opened) 
	VALUES (?, ?, ?, ?)`
	
	now := time.Now()
	_, err := db.conn.Exec(query, name, path, now, now)
	if err != nil {
		return fmt.Errorf("failed to add project: %w", err)
	}
	
	return nil
}

func (db *Database) GetProject(name string) (*Project, error) {
	query := `
	SELECT id, name, path, created_at, last_opened, description 
	FROM projects 
	WHERE name = ?`
	
	row := db.conn.QueryRow(query, name)
	
	var project Project
	var lastOpened sql.NullTime
	var description sql.NullString
	
	err := row.Scan(
		&project.ID,
		&project.Name,
		&project.Path,
		&project.CreatedAt,
		&lastOpened,
		&description,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project '%s' not found", name)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	if lastOpened.Valid {
		project.LastOpened = lastOpened.Time
	}
	if description.Valid {
		project.Description = description.String
	}
	
	return &project, nil
}

func (db *Database) ListProjects() ([]*Project, error) {
	query := `
	SELECT id, name, path, created_at, last_opened, description 
	FROM projects 
	ORDER BY last_opened DESC`
	
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()
	
	var projects []*Project
	
	for rows.Next() {
		var project Project
		var lastOpened sql.NullTime
		var description sql.NullString
		
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Path,
			&project.CreatedAt,
			&lastOpened,
			&description,
		)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		
		if lastOpened.Valid {
			project.LastOpened = lastOpened.Time
		}
		if description.Valid {
			project.Description = description.String
		}
		
		projects = append(projects, &project)
	}
	
	return projects, nil
}

func (db *Database) UpdateProjectLastOpened(projectID int) error {
	query := `UPDATE projects SET last_opened = ? WHERE id = ?`
	_, err := db.conn.Exec(query, time.Now(), projectID)
	if err != nil {
		return fmt.Errorf("failed to update project last_opened: %w", err)
	}
	return nil
}

func (db *Database) RemoveProject(name string) error {
	// First get the project ID
	project, err := db.GetProject(name)
	if err != nil {
		return err
	}
	
	// Remove from active_project if it's the active one
	_, err = db.conn.Exec("DELETE FROM active_project WHERE project_id = ?", project.ID)
	if err != nil {
		return fmt.Errorf("failed to remove active project reference: %w", err)
	}
	
	// Remove project settings
	_, err = db.conn.Exec("DELETE FROM project_settings WHERE project_id = ?", project.ID)
	if err != nil {
		return fmt.Errorf("failed to remove project settings: %w", err)
	}
	
	// Remove the project itself
	_, err = db.conn.Exec("DELETE FROM projects WHERE id = ?", project.ID)
	if err != nil {
		return fmt.Errorf("failed to remove project: %w", err)
	}
	
	return nil
}

func (db *Database) SetActiveProject(projectID int) error {
	// Use INSERT OR REPLACE to ensure only one active project
	query := `INSERT OR REPLACE INTO active_project (id, project_id) VALUES (1, ?)`
	_, err := db.conn.Exec(query, projectID)
	if err != nil {
		return fmt.Errorf("failed to set active project: %w", err)
	}
	return nil
}

func (db *Database) GetActiveProject() (*Project, error) {
	query := `
	SELECT p.id, p.name, p.path, p.created_at, p.last_opened, p.description 
	FROM projects p
	JOIN active_project ap ON p.id = ap.project_id
	WHERE ap.id = 1`
	
	row := db.conn.QueryRow(query)
	
	var project Project
	var lastOpened sql.NullTime
	var description sql.NullString
	
	err := row.Scan(
		&project.ID,
		&project.Name,
		&project.Path,
		&project.CreatedAt,
		&lastOpened,
		&description,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no active project set")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get active project: %w", err)
	}
	
	if lastOpened.Valid {
		project.LastOpened = lastOpened.Time
	}
	if description.Valid {
		project.Description = description.String
	}
	
	return &project, nil
}

func (db *Database) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}