package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

const (
	_ = iota
	Queued
	Started
	Success
	Failed
)

type DBConnect struct {
	db *sql.DB
}

// NewDBConnect is a factory method that returns a new db connection
func NewDBConnect(dsn string) (*DBConnect, error) {

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Validate DSN data
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DBConnect{db: db}, nil
}

// ValidAuth returns true if the API Key is valid for a request.
func (d *DBConnect) ValidAuth(key string) bool {
	var id int
	row := d.db.QueryRow("SELECT id FROM auth_tokens WHERE token = ?", key)
	err := row.Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		return false
	default:
		return true
	}
}

// AuthDeployEnv returns true if the API Key is valid to deploy to a given environment for a request.
func (d *DBConnect) AuthDeployEnv(key string, env string) bool {
	var id int
	row := d.db.QueryRow("SELECT ate.id as id "+
		"FROM auth_tokens_environments as ate "+
		"INNER JOIN auth_tokens as at "+
		"  ON ate.auth_token_id = at.id "+
		"  AND at.token = ? "+
		"INNER JOIN environments as e "+
		"  ON ate.environment_id = e.id "+
		"  AND e.name = ?", key, env)
	err := row.Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		return false
	default:
		return true
	}
}

// QueueDeploy inserts a fresh row into the log for a deployment run.
func (d *DBConnect) QueueDeploy(deployID string, environment string, imageName string, imageTag string) bool {
	msg := "Queued deploy."
	log := fmt.Sprintln(msg)
	result, err := d.db.Exec("INSERT INTO deploys (deploy_id, environment, image_name, image_tag, status, message, log, "+
		"updated_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())",
		deployID, environment, imageName, imageTag, Started, msg, log)
	if err != nil {
		return false
	}
	id, err := result.LastInsertId()
	if err != nil || id <= 0 {
		return false
	}
	return true
}

// UpdateDeploy updates the deploy row with information from the run.
func (d *DBConnect) UpdateDeploy(deployID string, status int, message string, log string) bool {
	result, err := d.db.Exec("UPDATE deploys "+
		"SET status = ?, "+
		"message = ?, "+
		"log = ?, "+
		"updated_at = NOW() "+
		"WHERE deploy_id = ?",
		status, message, log, deployID)
	if err != nil {
		return false
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return false
	}
	return true
}

// DeployStatus is used to return deploy status information from the database to the requester.
type DeployStatus struct {
	DeployID    string `json:"deployID"`    // UUID of teh deploy.
	Environment string `json:"environment"` // Environment serviced (development, qa etc.)
	ImageName   string `json:"imageName"`   // Docker image name.
	ImageTag    string `json:"imageTag"`    // Version tag of the image.
	Status      int    `json:"status"`      // The status ID of the result.
	Message     string `json:"message"`     // A user friendly message of what occurred.
	Log         string `json:"log"`         // The log of all steps run during the deploy.
	UpdatedAt   string `json:"updatedAt"`   // The create date and time of the deploy.
	CreatedAt   string `json:"createdAt"`   // The last update to this record.
}

// QueryDeploy returns the status of a deploy request.
func (d *DBConnect) QueryDeploy(deployID string) (*DeployStatus, error) {
	r := &DeployStatus{}
	row := d.db.QueryRow("SELECT deploy_id, environment, image_name, image_tag, status, message, log, "+
		"updated_at, created_at "+
		"FROM deploys WHERE deploy_id = ?", deployID)
	err := row.Scan(&r.DeployID, &r.Environment, &r.ImageName, &r.ImageTag, &r.Status, &r.Message, &r.Log,
		&r.UpdatedAt, &r.CreatedAt)
	switch {
	case err == sql.ErrNoRows:
		return nil, err
	case err != nil:
		return nil, err
	default:
		return r, nil
	}
}

// Close closes the connection(s) to the DB.
func (d *DBConnect) Close() bool {
	d.db.Close()
	return true
}
