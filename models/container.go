package models

import "time"
import "database/sql"

// Container represents an individual container that will contain items.
type Container struct {
	ID       int64     `json:"id"`
	User     User      `json:"-"`
	Name     string    `json:"name"`
	UUID     string    `json:"uuid"`
	Location Location  `json:"-"` // @todo update when location is implemented
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
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

// ContainerbyID retrieves a container by its primary ID
func ContainerbyID(db *sql.DB, ID int64) (Container, error) {
	var userID int64
	var locationID int64
	q := `
		select id, user_id, location_id, name, uuid, created, modified
		from containers
		where id = ?
	`
	var container Container
	err := db.QueryRow(q, ID).Scan(&container.ID, &userID, &locationID, &container.Name, &container.UUID, &container.Created, &container.Modified)
	if err != nil {
		return container, err
	}
	container.User, err = GetUserByID(db, userID)

	return container, err
}
