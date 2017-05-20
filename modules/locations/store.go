package locations

import (
	"database/sql"

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

type SortableField int

const (
	SortFieldModified SortableField = iota
	SortFieldName
	SortFieldContainerCount
)

var fields = [...]string{
	"modified",
	"name",
	"countainer_count",
}

func (field SortableField) String() string {
	return fields[field]
}

// GetSortBy will retrieve a SortBy object taylored for container queries
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
