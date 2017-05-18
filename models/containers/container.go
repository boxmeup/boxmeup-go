package containers

import (
	"time"

	"github.com/cjsaylor/boxmeup-go/models/locations"
	"github.com/cjsaylor/boxmeup-go/models/users"
)

// Container represents an individual container that will contain items.
type Container struct {
	ID                 int64              `json:"id"`
	User               users.User         `json:"-"`
	Name               string             `json:"name"`
	UUID               string             `json:"uuid"`
	Location           locations.Location `json:"-"` // @todo update when location is implemented
	ContainerItemCount int                `json:"container_item_count"`
	Created            time.Time          `json:"created"`
	Modified           time.Time          `json:"modified"`
}

// Containers is a group of containers
type Containers []Container
