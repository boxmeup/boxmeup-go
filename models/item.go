package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

// ContainerItem represents a single item in a container
type ContainerItem struct {
	ID        int64      `json:"id"`
	Container *Container `json:"-"`
	UUID      string     `json:"uuid"`
	Body      string     `json:"body"`
	Quantity  int        `json:"quantity"`
	Created   time.Time  `json:"created"`
	Modifed   time.Time  `json:"modifed"`
}

// ContainerItemStore persists and queries container items
type ContainerItemStore struct {
	DB *sql.DB
}

// GetSortBy will retrieve a SortBy object taylored for container queries
func (c *ContainerItemStore) GetSortBy(field string, direction SortType) SortBy {
	sortable := map[string]string{"modified": "modified", "body": "body", "quantity": "quantity"}
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

// Create will persist a given container item.
func (c *ContainerItemStore) Create(item *ContainerItem) error {
	q := `
		insert into container_items (container_id, uuid, body, quantity, created, modified)
		values(?, uuid(), ?, ?, now(), now())
	`
	res, err := c.DB.Exec(q, item.Container.ID, item.Body, item.Quantity)
	item.ID, _ = res.LastInsertId()

	return err
}

// Update a container item
func (c *ContainerItemStore) Update(item ContainerItem) error {
	if item.ID == 0 {
		return errors.New("can not update an item without it first being persisted")
	}
	q := `
		update container_items set body = ?, quantity = ?, modified = now()
		where id = ?
	`
	_, err := c.DB.Exec(q, item.Body, item.Quantity, item.ID)
	return err
}

// ContainerItems is a collection of container items.
type ContainerItems []ContainerItem

// ContainerItemsResponse is a response object that contains items and paginated meta.
type ContainerItemsResponse struct {
	Items         ContainerItems `json:"items"`
	PagedResponse PagedResponse  `json:"paged_response"`
}

// ByID retrieves an item by its ID
func (c *ContainerItemStore) ByID(ID int64) (ContainerItem, error) {
	q := `
		select id, container_id, uuid, body, quantity, created, modified
		from container_items
		where id = ?
	`
	item := ContainerItem{}
	var containerID int64
	err := c.DB.QueryRow(q, ID).Scan(&item.ID, &containerID, &item.UUID, &item.Body, &item.Quantity, &item.Created, &item.Modifed)
	if err != nil {
		return item, err
	}
	containerModel := ContainerStore{DB: c.DB}
	container, err := containerModel.ByID(containerID)
	if err == nil {
		item.Container = &container
	}
	return item, err
}

// GetContainerItems retrieves all items (paginated) from a container
func (c *ContainerItemStore) GetContainerItems(container *Container, sort SortBy, limit QueryLimit) (ContainerItemsResponse, error) {
	q := `
		select id, uuid, body, quantity, created, modified
		from container_items
		where container_id = ?
		order by %v %v
		limit %v offset %v
	`
	q = fmt.Sprintf(q, sort.Field, sort.Direction, limit.Limit, limit.Offset)
	rows, err := c.DB.Query(q, container.ID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	response := ContainerItemsResponse{}
	for rows.Next() {
		item := ContainerItem{}
		rows.Scan(&item.ID, &item.UUID, &item.Body, &item.Quantity, &item.Created, &item.Modifed)
		item.Container = container
		response.Items = append(response.Items, item)
	}
	response.PagedResponse.RequestTotal = len(response.Items)
	c.DB.QueryRow("select FOUND_ROWS()").Scan(&response.PagedResponse.Total)
	response.PagedResponse.CalculatePages(limit)
	return response, rows.Err()
}
