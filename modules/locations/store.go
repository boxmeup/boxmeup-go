package locations

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

// SortableField represents a field that is sortable
type SortableField int

const (
	// SortFieldID indicates ordering by ID
	SortFieldID SortableField = iota
	// SortFieldModified indicates ordering by modified
	SortFieldModified
	// SortFieldName indicates ordering by name
	SortFieldName
	// SortFieldContainerCount indicates ordering by container count
	SortFieldContainerCount
)

var fields = [...]string{
	"id",
	"modified",
	"name",
	"countainer_count",
}

// PagedResponse contains a group of locations and meta data for pagination
type PagedResponse struct {
	Locations     Locations            `json:"locations"`
	PagedResponse models.PagedResponse `json:"meta"`
}

func (field SortableField) String() string {
	return fields[field]
}

// SortableFieldByName sortable field by name
func (l *Store) SortableFieldByName(name string) (SortableField, error) {
	for i, field := range fields {
		if field == name {
			return SortableField(i), nil
		}
	}
	return -1, errors.New("sortable field not found")
}

// GetSortBy will retrieve a SortBy object taylored for location queries
func (l *Store) GetSortBy(sortField SortableField, direction models.SortType) models.SortBy {
	var sort models.SortBy
	if field := fields[sortField]; field != "" {
		sort.Field = field
	} else {
		sort.Field = SortFieldModified.String()
	}
	if direction == models.ASC {
		sort.Direction = models.ASC
	} else {
		sort.Direction = models.DSC
	}
	return sort
}

// Create a location entry
func (l *Store) Create(location *Location) error {
	q := `
		insert into locations (user_id, uuid, name, is_mappable, address, created, modified)
		values (?, uuid(), ?, ?, ?, now(), now())
	`
	res, err := l.DB.Exec(q, location.User.ID, location.Name, location.Address != "", location.Address)
	location.ID, _ = res.LastInsertId()
	return err
}

// Update will update details of the provided location
func (l *Store) Update(location *Location) error {
	if location.ID == 0 {
		return errors.New("location must already be stored")
	}
	q := `
		update locations set name = ?, address = ?, modified = now() where id = ?
	`
	_, err := l.DB.Exec(q, location.Name, location.Address, location.ID)
	return err
}

// Delete will remove a location by ID.
func (l *Store) Delete(ID int64) error {
	q := "delete from locations where ID = ?"
	_, err := l.DB.Exec(q, ID)
	return err
}

// ByID will return a location by its identifier.
func (l *Store) ByID(ID int64) (Location, error) {
	q := `
		select id, user_id, uuid, name, address, container_count, created, modified
		from locations where id = ?
	`
	var location Location
	var userID int64
	err := l.DB.QueryRow(q, ID).Scan(
		&location.ID,
		&userID,
		&location.UUID,
		&location.Name,
		&location.Address,
		&location.ContainerCount,
		&location.Created,
		&location.Modified)
	if err == nil {
		location.User, err = users.NewStore(l.DB).ByID(userID)
	}
	return location, err
}

// UserLocations will get all containers belonging to a user
func (l *Store) UserLocations(user users.User, sort models.SortBy, limit models.QueryLimit) (PagedResponse, error) {
	q := `
		select SQL_CALC_FOUND_ROWS id, uuid, name, address, container_count, created, modified
		from locations
		where user_id = ?
		order by %v %v
		limit %v offset %v
	`
	q = fmt.Sprintf(q, sort.Field, sort.Direction, limit.Limit, limit.Offset)
	rows, err := l.DB.Query(q, user.ID)
	if err != nil {
		log.Fatal(err)
	}
	response := PagedResponse{}
	defer rows.Close()
	for rows.Next() {
		location := Location{}
		rows.Scan(
			&location.ID,
			&location.UUID,
			&location.Name,
			&location.Address,
			&location.ContainerCount,
			&location.Created,
			&location.Modified)
		response.Locations = append(response.Locations, location)
	}
	response.PagedResponse.RequestTotal = len(response.Locations)
	l.DB.QueryRow("select FOUND_ROWS()").Scan(&response.PagedResponse.Total)
	response.PagedResponse.CalculatePages(limit)
	return response, rows.Err()
}
