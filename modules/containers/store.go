package containers

import (
	"database/sql"
	"fmt"
	"log"

	"errors"

	"github.com/cjsaylor/boxmeup-go/modules/models"
	"github.com/cjsaylor/boxmeup-go/modules/users"
)

// QueryLimit is the maximum number of container results per page.
const QueryLimit = 20

// NewStore constructs a storage interface for containers.
func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

// Store helps store and retrieve
type Store struct {
	DB *sql.DB
}

// GetSortBy will retrieve a SortBy object taylored for container queries
func (c *Store) GetSortBy(field string, direction models.SortType) models.SortBy {
	sortable := map[string]string{"modified": "modified", "name": "name"}
	var sort models.SortBy
	if _, ok := sortable[field]; ok {
		sort.Field = field
	} else {
		sort.Field = "modified"
	}
	if direction == models.ASC {
		sort.Direction = models.ASC
	} else {
		sort.Direction = models.DSC
	}

	return sort
}

// Create persists a container to the database
func (c *Store) Create(container *Container) error {
	q := `
		insert into containers (user_id, location_id, name, uuid, created, modified)
		values (?, ?, ?, uuid(), now(), now())
	`
	res, err := c.DB.Exec(q, container.User.ID, container.Location.ID, container.Name)
	container.ID, _ = res.LastInsertId()

	return err
}

// Update a container
// @todo add support for location updating
func (c *Store) Update(container *Container) error {
	if container.ID == 0 {
		return errors.New("can not update a container without it first being persisted")
	}
	q := `
		update container set name = ?, modified = now()
		where id = ?
	`
	_, err := c.DB.Exec(q, container.Name, container.ID)
	return err
}

// Delete will remove a container by its ID.
// Note that due to the FK constrant set to cascade on deletion, this will
// delete all the related items as well.
func (c *Store) Delete(ID int64) error {
	// Note, the FK has cascade deletion, so this will delete the items as well.
	q := "delete from containers where id = ?"
	_, err := c.DB.Exec(q, ID)
	return err
}

// ByID retrieves a container by its primary ID
func (c *Store) ByID(ID int64) (Container, error) {
	var userID int64
	var locationID int64
	q := `
		select id, user_id, location_id, name, uuid, container_item_count, created, modified
		from containers
		where id = ?
	`
	var container Container
	err := c.DB.QueryRow(q, ID).Scan(
		&container.ID,
		&userID,
		&locationID,
		&container.Name,
		&container.UUID,
		&container.ContainerItemCount,
		&container.Created,
		&container.Modified)
	if err != nil {
		return container, err
	}
	userModel := users.NewStore(c.DB)
	container.User, err = userModel.ByID(userID)

	return container, err
}

// PagedResponse contains a group of containers and meta data for pagination
type PagedResponse struct {
	Containers    Containers           `json:"containers"`
	PagedResponse models.PagedResponse `json:"meta"`
}

// UserContainers will get all containers belonging to a user
func (c *Store) UserContainers(user users.User, sort models.SortBy, limit models.QueryLimit) (PagedResponse, error) {
	q := `
		select SQL_CALC_FOUND_ROWS id, location_id, name, uuid, container_item_count, created, modified
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
	response := PagedResponse{}
	defer rows.Close()
	var locationID int64
	for rows.Next() {
		container := Container{}
		rows.Scan(
			&container.ID,
			&locationID,
			&container.Name,
			&container.UUID,
			&container.ContainerItemCount,
			&container.Created,
			&container.Modified)
		response.Containers = append(response.Containers, container)
	}
	response.PagedResponse.RequestTotal = len(response.Containers)
	c.DB.QueryRow("select FOUND_ROWS()").Scan(&response.PagedResponse.Total)
	response.PagedResponse.CalculatePages(limit)
	return response, rows.Err()
}
