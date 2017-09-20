package locations

import (
	"time"

	"github.com/cjsaylor/boxmeup-go/modules/users"
)

// Location structure
type Location struct {
	ID             int64      `json:"id"`
	User           users.User `json:"-"`
	UUID           string     `json:"uuid"`
	Name           string     `json:"name"`
	Address        string     `json:"address"`
	ContainerCount int        `json:"container_count"`
	Created        time.Time  `json:"created"`
	Modified       time.Time  `json:"modified"`
}

// Locations group of locations
type Locations []Location

type LocationFilter struct {
	User                  users.User
	ContainerID           int64
	IsAttachedToContainer bool
}
