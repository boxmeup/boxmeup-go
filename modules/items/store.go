package items

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/cjsaylor/boxmeup-go/modules/containers"
	"github.com/cjsaylor/boxmeup-go/modules/models"
)

// Store persists and queries container items
type Store struct {
	DB *sql.DB
}

// NewStore constructs a storage interface for items.
func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

// GetSortBy will retrieve a SortBy object taylored for container queries
func (c *Store) GetSortBy(field string, direction models.SortType) models.SortBy {
	sortable := map[string]string{"modified": "modified", "body": "body", "quantity": "quantity"}
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

// Create will persist a given container item.
func (c *Store) Create(item *ContainerItem) error {
	q := `
		insert into container_items (container_id, uuid, body, quantity, created, modified)
		values(?, uuid(), ?, ?, now(), now())
	`
	tx, _ := c.DB.Begin()
	res, err := tx.Exec(q, item.Container.ID, item.Body, item.Quantity)
	item.ID, _ = res.LastInsertId()
	if err == nil {
		err = updateContainerItemCount(tx, item.Container.ID)
	}
	if err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}

	return err
}

// Update a container item
func (c *Store) Update(item ContainerItem) error {
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

// @todo determine if this should be the number of "rows" or if it should be based on quantity
// Also consider moving this to a MySQL trigger
func updateContainerItemCount(tx *sql.Tx, containerID int64) error {
	q := `
		update containers
		set container_item_count = (
			select count(*) from container_items where container_id = ?
		), modified = now()
		where id = ?
	`
	_, err := tx.Exec(q, containerID, containerID)
	return err
}

// Delete removes an item from a container
func (c *Store) Delete(item ContainerItem) error {
	q := "delete from container_items where id = ?"
	tx, _ := c.DB.Begin()
	_, err := tx.Exec(q, item.ID)
	if err == nil {
		err = updateContainerItemCount(tx, item.Container.ID)
	}
	if err == nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
	return err
}

// PagedResponse is a response object that contains items and paginated meta.
type PagedResponse struct {
	Items         ContainerItems       `json:"items"`
	PagedResponse models.PagedResponse `json:"paged_response"`
}

func (r *PagedResponse) getItemIDMap() map[int64]*ContainerItem {
	mappedItems := make(map[int64]*ContainerItem)
	for index, v := range r.Items {
		mappedItems[v.ID] = &r.Items[index]
	}
	return mappedItems
}

// ByID retrieves an item by its ID
func (c *Store) ByID(ID int64) (ContainerItem, error) {
	q := `
		select id, container_id, uuid, body, quantity, created, modified
		from container_items
		where id = ?
	`
	item := ContainerItem{}
	var containerID int64
	err := c.DB.QueryRow(q, ID).Scan(&item.ID, &containerID, &item.UUID, &item.Body, &item.Quantity, &item.Created, &item.Modified)
	if err != nil {
		return item, err
	}
	container, err := containers.NewStore(c.DB).ByID(containerID)
	if err == nil {
		item.Container = &container
	}
	return item, err
}

// GetContainerItems retrieves all items (paginated) from a container
func (c *Store) GetContainerItems(container *containers.Container, sort models.SortBy, limit models.QueryLimit) (PagedResponse, error) {
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
	response := PagedResponse{}
	for rows.Next() {
		item := ContainerItem{}
		rows.Scan(&item.ID, &item.UUID, &item.Body, &item.Quantity, &item.Created, &item.Modified)
		item.Container = container
		response.Items = append(response.Items, item)
	}
	response.PagedResponse.RequestTotal = len(response.Items)
	c.DB.QueryRow("select FOUND_ROWS()").Scan(&response.PagedResponse.Total)
	response.PagedResponse.CalculatePages(limit)
	return response, rows.Err()
}

func (c *Store) SearchItems(userID int64, term string, sort models.SortBy, limit models.QueryLimit) (PagedResponse, error) {
	q := `
		select ci.id, container_id, ci.uuid, body, quantity, ci.created, ci.modified
		from container_items ci
		inner join containers c on c.id = ci.container_id and c.user_id = ?
		where body like concat('%%', ?, '%%')
		order by %v %v
		limit %v offset %v
	`
	q = fmt.Sprintf(q, sort.Field, sort.Direction, limit.Limit, limit.Offset)
	rows, err := c.DB.Query(q, userID, term)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	response := PagedResponse{}
	containerIDs := make(map[int64]int64)
	for rows.Next() {
		item := ContainerItem{}
		var containerID int64
		rows.Scan(&item.ID, &containerID, &item.UUID, &item.Body, &item.Quantity, &item.Created, &item.Modified)
		containerIDs[item.ID] = containerID
		response.Items = append(response.Items, item)
	}
	response.PagedResponse.RequestTotal = len(response.Items)
	c.DB.QueryRow("select FOUND_ROWS()").Scan(&response.PagedResponse.Total)
	var wg sync.WaitGroup
	wg.Add(len(containerIDs))
	itemMap := response.getItemIDMap()
	containerModel := containers.NewStore(c.DB)
	for k, v := range containerIDs {
		go func(containerID int64, item *ContainerItem) {
			defer wg.Done()
			// @todo Instead of doing individual queries, make a query in containers to accept
			// an array of IDs and do a single query
			// For the time being this is fine because we limit the maximum results to QueryLimit (20)
			container, _ := containerModel.ByID(containerID)
			item.Container = &container
		}(v, itemMap[k])
	}
	response.PagedResponse.CalculatePages(limit)
	wg.Wait()
	return response, rows.Err()

}
