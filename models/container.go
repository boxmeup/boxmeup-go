package models

import "time"
import "database/sql"

// Container represents an individual container that will contain items.
type Container struct {
	ID       int64     `json:"id,omitempty"`
	User     User      `json:"-,omitempty"`
	Name     string    `json:"name,omitempty"`
	Location Location  `json:"location,omitempty"`
	Created  time.Time `json:"created,omitempty"`
	Modified time.Time `json:"modified,omitempty"`
}

// CreateContainer persists a container to the database
func CreateContainer(db *sql.DB, container *Container) error {
	q := `
		insert into containers (user_id, location_id, name, uuid, created, modified)
		values (?, ?, ?, uuid(), now(), now())
	`

	res, err := db.Exec(q, container.User.ID, container.Location.ID, container.Name)
	container.ID, _ = res.LastInsertId()

	return err
}
