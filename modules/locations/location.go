package locations

import (
	"time"

	"github.com/cjsaylor/boxmeup-go/modules/users"
)

// Location structure
type Location struct {
	ID             int64
	User           users.User
	UUID           string
	Name           string
	Address        string
	ContainerCount int
	Created        time.Time
	Modified       time.Time
}
