package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

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

// ContainerQueryLimit is the maximum number of container results per page.
const ContainerQueryLimit = 20

// Containers is a group of containers
type Containers []Container

// ContainerStore helps store and retrieve
type ContainerStore struct {
	DB *sql.DB
}

// GetSortBy will retrieve a SortBy object taylored for container queries
func (c *ContainerStore) GetSortBy(field string, direction SortType) SortBy {
	sortable := map[string]string{"modified": "modified", "name": "name"}
	var sort SortBy
	if _, ok := sortable[field]; ok {
		sort.Field = field
	} else {
		sort.Field = "modified"
	}
	if direction == ASC {
		sort.Direction = ASC
	} else {
		sort.Direction = DSC
	}

	return sort
}

// Create persists a container to the database
func (c *ContainerStore) Create(container *Container) error {
	q := `
		insert into containers (user_id, location_id, name, uuid, created, modified)
		values (?, ?, ?, uuid(), now(), now())
	`
	res, err := c.DB.Exec(q, container.User.ID, container.Location.ID, container.Name)
	container.ID, _ = res.LastInsertId()

	return err
}

// ByID retrieves a container by its primary ID
func (c *ContainerStore) ByID(ID int64) (Container, error) {
	var userID int64
	var locationID int64
	q := `
		select id, user_id, location_id, name, uuid, created, modified
		from containers
		where id = ?
	`
	var container Container
	err := c.DB.QueryRow(q, ID).Scan(&container.ID, &userID, &locationID, &container.Name, &container.UUID, &container.Created, &container.Modified)
	if err != nil {
		return container, err
	}
	userModel := UserStore{DB: c.DB}
	container.User, err = userModel.ByID(userID)

	return container, err
}

// ContainersResponse contains a group of containers and meta data for pagination
type ContainersResponse struct {
	Containers    Containers    `json:"containers"`
	PagedResponse PagedResponse `json:"meta"`
}

// UserContainers will get all containers belonging to a user
func (c *ContainerStore) UserContainers(user User, sort SortBy, limit QueryLimit) (ContainersResponse, error) {
	q := `
		select SQL_CALC_FOUND_ROWS id, location_id, name, uuid, created, modified
		from containers
		where user_id = ?
		order by %v %v
		limit %v offset %v
	`
	q = fmt.Sprintf(q, sort.Field, sort.Direction, limit.Limit, limit.Offset)
	rows, err := c.DB.Query(q, user.ID)
	if err != nil {
		log.Fatal(err)
	}
	response := ContainersResponse{}
	defer rows.Close()
	var locationID int64
	for rows.Next() {
		container := Container{}
		rows.Scan(&container.ID, &locationID, &container.Name, &container.UUID, &container.Created, &container.Modified)
		response.Containers = append(response.Containers, container)
	}
	response.PagedResponse.RequestTotal = len(response.Containers)
	c.DB.QueryRow("select FOUND_ROWS()").Scan(&response.PagedResponse.Total)
	response.PagedResponse.CalculatePages(limit)
	return response, rows.Err()
}
