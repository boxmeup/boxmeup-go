package items

import (
	"time"

	"github.com/cjsaylor/boxmeup-go/modules/containers"
)

// ContainerItem represents a single item in a container
type ContainerItem struct {
	ID        int64                 `json:"id"`
	Container *containers.Container `json:"-"`
	UUID      string                `json:"uuid"`
	Body      string                `json:"body"`
	Quantity  int                   `json:"quantity"`
	Created   time.Time             `json:"created"`
	Modified  time.Time             `json:"modifed"`
}

// ContainerItems is a collection of container items.
type ContainerItems []ContainerItem
