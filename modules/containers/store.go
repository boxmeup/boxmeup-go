package containers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/cjsaylor/boxmeup-go/modules/locations"
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
func (c *Store) Create(record *ContainerRecord) error {
	if record.Name == "" {
		return errors.New("Container must have a name")
	}
	q := `
		insert into containers (user_id, location_id, name, uuid, created, modified)
		values (?, ?, ?, uuid(), now(), now())
	`
	tx, _ := c.DB.Begin()
	res, err := tx.Exec(q, record.userID, record.locationID, record.Name)
	if err == nil && record.locationID > 0 {
		err = updateContainerCount(tx, record.locationID)
	}
	if err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
	record.ID, _ = res.LastInsertId()

	return err
}

// Update a container
func (c *Store) Update(record *ContainerRecord) error {
	if record.ID == 0 {
		return errors.New("can not update a container without it first being persisted")
	}
	if record.Name == "" {
		return errors.New("containers must have a name")
	}
	q := `
		update containers set name = ?, location_id = ?, modified = now()
		where id = ?
	`
	tx, _ := c.DB.Begin()
	_, err := tx.Exec(q, record.Name, record.locationID, record.ID)
	if err == nil {
		if record.locationID > 0 {
			err = updateContainerCount(tx, record.locationID)
		}
		if err == nil && record.oldLocationID > 0 {
			err = updateContainerCount(tx, record.oldLocationID)
		}
	}
	if err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
	return err
}

// Delete will remove a container by its ID.
// Note that due to the FK constrant set to cascade on deletion, this will
// delete all the related items as well.
func (c *Store) Delete(ID int64) error {
	// Note, the FK has cascade deletion, so this will delete the items as well.
	q := "delete from containers where id = ?"
	tx, _ := c.DB.Begin()
	_, err := tx.Exec(q, ID)
	if err != nil {
		err = updateContainerCount(tx, ID)
	}
	if err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
	return err
}

// @todo consider moving this to a MySQL trigger
func updateContainerCount(tx *sql.Tx, locationID int64) error {
	q := `
		update locations
		set container_count = (
			select count(*) from containers where location_id = ?
		), modified = now()
		where id = ?
	`
	_, err := tx.Exec(q, locationID, locationID)
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
	var wg sync.WaitGroup
	wg.Add(2)
	go func(userID int64, container *Container) {
		defer wg.Done()
		container.User, err = users.NewStore(c.DB).ByID(userID)
	}(userID, &container)
	go func(locationID int64, container *Container) {
		defer wg.Done()
		location, _ := locations.NewStore(c.DB).ByID(locationID)
		if location.ID > 0 {
			container.Location = &location
		}
	}(locationID, &container)

	wg.Wait()

	return container, err
}

func (r *PagedResponse) getContainerIDMap() map[int64]*Container {
	mappedContainers := make(map[int64]*Container)
	for index, v := range r.Containers {
		mappedContainers[v.ID] = &r.Containers[index]
	}
	return mappedContainers
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
	locationIDs := make(map[int64]int64)
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
		if locationID > 0 {
			locationIDs[container.ID] = locationID
		}
		response.Containers = append(response.Containers, container)
	}
	response.PagedResponse.RequestTotal = len(response.Containers)
	c.DB.QueryRow("select FOUND_ROWS()").Scan(&response.PagedResponse.Total)
	var wg sync.WaitGroup
	wg.Add(len(locationIDs))
	containerMap := response.getContainerIDMap()
	for k, v := range locationIDs {
		go func(locationID int64, container *Container) {
			defer wg.Done()
			// @todo Instead of doing individual queries, make a query in locations to accept
			// an array of IDs and do a single query
			// For the time being this is fine because we limit the maximum results to QueryLimit (20)
			location, _ := locations.NewStore(c.DB).ByID(locationID)
			if location.ID > 0 {
				container.Location = &location
			}
		}(v, containerMap[k])
	}
	response.PagedResponse.CalculatePages(limit)
	wg.Wait()
	return response, rows.Err()
}
