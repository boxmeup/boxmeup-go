package containers

import (
	"time"

	"github.com/cjsaylor/boxmeup-go/modules/locations"
	"github.com/cjsaylor/boxmeup-go/modules/models"
	"github.com/cjsaylor/boxmeup-go/modules/users"
)

// Container represents an individual container that will contain items.
type Container struct {
	ID                 int64               `json:"id"`
	User               users.User          `json:"-"`
	Name               string              `json:"name"`
	UUID               string              `json:"uuid"`
	Location           *locations.Location `json:"location"`
	ContainerItemCount int                 `json:"container_item_count"`
	Created            time.Time           `json:"created"`
	Modified           time.Time           `json:"modified"`
}

type ContainerRecord struct {
	ID            int64
	userID        int64
	locationID    int64
	oldLocationID int64
	Name          string
}

type ContainerFilter struct {
	User        users.User
	LocationIDs []int64
}

func (f *ContainerFilter) GenericLocationIDList() []interface{} {
	list := make([]interface{}, len(f.LocationIDs))
	for i, val := range f.LocationIDs {
		list[i] = val
	}
	return list
}

// PagedResponse contains a group of containers and meta data for pagination
type PagedResponse struct {
	Containers    Containers           `json:"containers"`
	PagedResponse models.PagedResponse `json:"meta"`
}

func NewRecord(user *users.User) ContainerRecord {
	record := ContainerRecord{}
	record.SetUser(user)
	return record
}

func (r *ContainerRecord) SetLocation(location *locations.Location) *ContainerRecord {
	r.oldLocationID = r.locationID
	if location == nil {
		r.locationID = 0
	} else {
		r.locationID = location.ID
	}
	return r
}

func (r *ContainerRecord) SetUser(user *users.User) *ContainerRecord {
	r.userID = user.ID
	return r
}

// Containers is a group of containers
type Containers []Container

func (c *Container) ToRecord() ContainerRecord {
	record := NewRecord(&c.User)
	if c.ID > 0 {
		record.ID = c.ID
	}
	if c.Location != nil {
		record.SetLocation(c.Location)
	}
	return record
}
